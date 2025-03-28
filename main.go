package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ghproxy/api"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/middleware/loggin"
	"ghproxy/middleware/timing"
	"ghproxy/proxy"
	"ghproxy/rate"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	cfg         *config.Config
	router      *gin.Engine
	configfile  = "/data/ghproxy/config/config.toml"
	cfgfile     string
	version     string
	dev         string
	runMode     string
	limiter     *rate.RateLimiter
	iplimiter   *rate.IPRateLimiter
	showVersion bool
	showHelp    bool
)

var (
	//go:embed pages/*
	pagesFS embed.FS
)

var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func readFlag() {
	flag.StringVar(&cfgfile, "cfg", configfile, "config file path")
	flag.BoolVar(&showVersion, "v", false, "show version and exit")   // 添加-v标志
	flag.BoolVar(&showHelp, "h", false, "show help message and exit") // 添加-h标志

	// 捕获未定义的 flag
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nInvalid flags:")

		// 检查未定义的flags
		invalidFlags := []string{}
		for _, arg := range os.Args[1:] {
			if arg[0] == '-' && arg != "-h" && arg != "-v" { // 检查是否是flag, 排除 -h 和 -v
				defined := false
				flag.VisitAll(func(f *flag.Flag) {
					if "-"+f.Name == arg {
						defined = true
					}
				})
				if !defined {
					invalidFlags = append(invalidFlags, arg)
				}
			}
		}
		for _, flag := range invalidFlags {
			fmt.Fprintf(os.Stderr, "  %s\n", flag)
		}
		if len(invalidFlags) > 0 {
			os.Exit(2) // 使用非零状态码退出，表示有错误
		}

	}
}

func loadConfig() {
	var err error
	cfg, err = config.LoadConfig(cfgfile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		// 如果配置文件加载失败，也显示帮助信息并退出
		flag.Usage()
		os.Exit(1)
	}
	if cfg != nil && cfg.Server.Debug { // 确保 cfg 不为 nil
		fmt.Println("Config File Path: ", cfgfile)
		fmt.Printf("Loaded config: %v\n", cfg)
	}
}

func setupLogger(cfg *config.Config) {
	var err error
	err = logger.Init(cfg.Log.LogFilePath, cfg.Log.MaxLogSize)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
	}
	err = logger.SetLogLevel(cfg.Log.Level)
	if err != nil {
		fmt.Printf("Logger Level Error: %v\n", err)
	}
	fmt.Printf("Log Level: %s\n", cfg.Log.Level)
	logDebug("Config File Path: ", cfgfile)
	logDebug("Loaded config: %v\n", cfg)
	logInfo("Init Completed")
}

func loadlist(cfg *config.Config) {
	auth.Init(cfg)
}

func setupApi(cfg *config.Config, router *gin.Engine, version string) {
	api.InitHandleRouter(cfg, router, version)
}

func setupRateLimit(cfg *config.Config) {
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RateMethod == "ip" {
			iplimiter = rate.NewIPRateLimiter(cfg.RateLimit.RatePerMinute, cfg.RateLimit.Burst, 1*time.Minute)
		} else if cfg.RateLimit.RateMethod == "total" {
			limiter = rate.New(cfg.RateLimit.RatePerMinute, cfg.RateLimit.Burst, 1*time.Minute)
		} else {
			logError("Invalid RateLimit Method: %s", cfg.RateLimit.RateMethod)
		}
	}
}

func InitReq(cfg *config.Config) {
	proxy.InitReq(cfg)
}

// loadEmbeddedPages 加载嵌入式页面资源
func loadEmbeddedPages(cfg *config.Config) (fs.FS, fs.FS, error) {
	var pages fs.FS
	var err error
	switch cfg.Pages.Theme {
	case "bootstrap":
		pages, err = fs.Sub(pagesFS, "pages/bootstrap")
	case "nebula":
		pages, err = fs.Sub(pagesFS, "pages/nebula")
	case "design":
		pages, err = fs.Sub(pagesFS, "pages/design")
	case "metro":
		pages, err = fs.Sub(pagesFS, "pages/metro")
	case "classic":
		pages, err = fs.Sub(pagesFS, "pages/classic")
	case "mino":
		pages, err = fs.Sub(pagesFS, "pages/mino")
	default:
		pages, err = fs.Sub(pagesFS, "pages/bootstrap") // 默认主题
		logWarning("Invalid Pages Theme: %s, using default theme 'bootstrap'", cfg.Pages.Theme)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to load embedded pages: %w", err)
	}

	var assets fs.FS
	assets, err = fs.Sub(pagesFS, "pages/assets")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load embedded assets: %w", err)
	}
	return pages, assets, nil
}

// setupPages 设置页面路由
func setupPages(cfg *config.Config, router *gin.Engine) {
	switch cfg.Pages.Mode {
	case "internal":
		// 加载嵌入式资源
		pages, assets, err := loadEmbeddedPages(cfg)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			return
		}

		// 设置嵌入式资源路由
		router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/bootstrap.min.css", gin.WrapH(http.FileServer(http.FS(assets))))
		router.GET("/bootstrap.bundle.min.js", gin.WrapH(http.FileServer(http.FS(assets))))

	case "external":
		// 设置外部资源路径
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
		stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)
		bootstrapPath := fmt.Sprintf("%s/bootstrap.min.css", cfg.Pages.StaticDir)
		bootstrapBundlePath := fmt.Sprintf("%s/bootstrap.bundle.min.js", cfg.Pages.StaticDir)

		// 设置外部资源路由
		router.GET("/", func(c *gin.Context) {
			c.File(indexPagePath)
			logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
		router.StaticFile("/favicon.ico", faviconPath)
		router.StaticFile("/script.js", javascriptsPath)
		router.StaticFile("/style.css", stylesheetsPath)
		router.StaticFile("/bootstrap.min.css", bootstrapPath)
		router.StaticFile("/bootstrap.bundle.min.js", bootstrapBundlePath)

		//router.StaticFile("/bootstrap.min.css", bootstrapPath)

	default:
		// 处理无效的Pages Mode
		logWarning("Invalid Pages Mode: %s, using default embedded theme", cfg.Pages.Mode)

		// 加载嵌入式资源
		pages, assets, err := loadEmbeddedPages(cfg)
		if err != nil {
			logError("Failed when processing pages: %s", err)
			return
		}
		// 设置嵌入式资源路由
		router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/bootstrap.min.css", gin.WrapH(http.FileServer(http.FS(assets))))
		router.GET("/bootstrap.bundle.min.js", gin.WrapH(http.FileServer(http.FS(assets))))
	}
}

func init() {
	readFlag()
	flag.Parse()

	// 如果设置了 -h，则显示帮助信息并退出
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 如果设置了 -v，则显示版本号并退出
	if showVersion {
		fmt.Printf("GHProxy Version: %s \n", version)
		os.Exit(0)
	}

	loadConfig()
	setupLogger(cfg)
	InitReq(cfg)
	loadlist(cfg)
	setupRateLimit(cfg)

	if cfg.Server.Debug {
		dev = "true"
		version = "dev"
	}
	if dev == "true" {
		gin.SetMode(gin.DebugMode)
		runMode = "dev"
	} else {
		gin.SetMode(gin.ReleaseMode)
		runMode = "release"
	}

	logDebug("Run Mode: %s", runMode)

	gin.LoggerWithWriter(io.Discard)
	router = gin.New()

	// 添加recovery中间件
	router.Use(gin.Recovery())

	// 添加log中间件
	router.Use(loggin.Middleware())

	// 添加计时中间件
	router.Use(timing.Middleware())

	if cfg.Server.H2C {
		router.UseH2C = true
	}

	setupApi(cfg, router, version)

	setupPages(cfg, router)

	// 1. GitHub Releases/Archive - Use distinct path segments for type
	router.GET("/github.com/:username/:repo/releases/*filepath", func(c *gin.Context) { // Distinct path for releases
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	router.GET("/github.com/:username/:repo/archive/*filepath", func(c *gin.Context) { // Distinct path for archive
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	// 2. GitHub Blob/Raw - Use distinct path segments for type
	router.GET("/github.com/:username/:repo/blob/*filepath", func(c *gin.Context) { // Distinct path for blob
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	router.GET("/github.com/:username/:repo/raw/*filepath", func(c *gin.Context) { // Distinct path for raw
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	router.GET("/github.com/:username/:repo/info/*filepath", func(c *gin.Context) { // Distinct path for info
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})
	router.GET("/github.com/:username/:repo/git-upload-pack", func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	// 4. Raw GitHubusercontent - Keep as is (assuming it's distinct enough)
	router.GET("/raw.githubusercontent.com/:username/:repo/*filepath", func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	// 5. Gist GitHubusercontent - Keep as is (assuming it's distinct enough)
	router.GET("/gist.githubusercontent.com/:username/*filepath", func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	// 6. GitHub API Repos - Keep as is (assuming it's distinct enough)
	router.GET("/api.github.com/repos/:username/:repo/*filepath", func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	router.NoRoute(func(c *gin.Context) {
		logInfo(c.Request.URL.Path)
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	fmt.Printf("GHProxy Version: %s\n", version)
	fmt.Printf("A Go Based High-Performance Github Proxy \n")
	fmt.Printf("Made by WJQSERVER-STUDIO\n")
}

func main() {
	if showVersion || showHelp {
		return
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	/*
		go func() {
			err := router.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
			if err != nil {
				logError("Failed to start server: %v\n", err)
			}
		}()
	*/

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logError("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		logError("Server forced to shutdown: %v\n", err)
	}
	defer cancel()
	logger.Close()
	fmt.Println("Program Exit")
}
