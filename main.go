package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"ghproxy/api"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/middleware/loggin"
	"ghproxy/proxy"
	"ghproxy/rate"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	//"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/hertz-contrib/http2/factory"
)

var (
	cfg         *config.Config
	r           *server.Hertz
	configfile  = "/data/ghproxy/config/config.toml"
	hertZfile   *os.File
	cfgfile     string
	version     string
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
	flag.StringVar(&cfgfile, "c", configfile, "config file path")
	flag.Func("cfg", "exit", func(s string) error {

		// 被弃用的flag, fail退出
		fmt.Printf("\n")
		fmt.Println("[ERROR] cfg flag is deprecated, please use -c instead")
		fmt.Printf("\n")
		flag.Usage()
		os.Exit(2)
		return nil
	})
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
			os.Exit(2)
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
		os.Exit(1)
	}
	err = logger.SetLogLevel(cfg.Log.Level)
	if err != nil {
		fmt.Printf("Logger Level Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Log Level: %s\n", cfg.Log.Level)
	logDebug("Config File Path: ", cfgfile)
	logDebug("Loaded config: %v\n", cfg)
	logInfo("Logger Initialized Successfully")
}

func setupHertZLogger(cfg *config.Config) {
	var err error

	if cfg.Log.HertZLogPath != "" {
		hertZfile, err = os.OpenFile(cfg.Log.HertZLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			hlog.SetOutput(os.Stdout)
			logWarning("Failed to open hertz log file: %v", err)
		} else {
			hlog.SetOutput(hertZfile)
		}
		hlog.SetLevel(hlog.LevelInfo)
	}

}

func setMemLimit(cfg *config.Config) {
	if cfg.Server.MemLimit > 0 {
		debug.SetMemoryLimit((cfg.Server.MemLimit) * 1024 * 1024)
		logInfo("Set Memory Limit to %d MB", cfg.Server.MemLimit)
	}
}

func loadlist(cfg *config.Config) {
	auth.Init(cfg)
}

func setupApi(cfg *config.Config, r *server.Hertz, version string) {
	api.InitHandleRouter(cfg, r, version)
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
	return pages, assets, nil
}

// setupPages 设置页面路由
func setupPages(cfg *config.Config, r *server.Hertz) {
	switch cfg.Pages.Mode {
	case "internal":
		err := setInternalRoute(cfg, r)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			fmt.Println(err.Error())
			return
		}

	case "external":
		// 设置外部资源路径
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
		stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)
		bootstrapPath := fmt.Sprintf("%s/bootstrap.min.css", cfg.Pages.StaticDir)
		bootstrapBundlePath := fmt.Sprintf("%s/bootstrap.bundle.min.js", cfg.Pages.StaticDir)

		// 设置外部资源路由
		r.StaticFile("/", indexPagePath)
		r.StaticFile("/favicon.ico", faviconPath)
		r.StaticFile("/script.js", javascriptsPath)
		r.StaticFile("/style.css", stylesheetsPath)
		r.StaticFile("/bootstrap.min.css", bootstrapPath)
		r.StaticFile("/bootstrap.bundle.min.js", bootstrapBundlePath)
		//router.StaticFile("/bootstrap.min.css", bootstrapPath)

	default:
		// 处理无效的Pages Mode
		logWarning("Invalid Pages Mode: %s, using default embedded theme", cfg.Pages.Mode)

		err := setInternalRoute(cfg, r)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			fmt.Println(err.Error())
			return
		}

	}
}

func setInternalRoute(cfg *config.Config, r *server.Hertz) error {

	// 加载嵌入式资源
	pages, assets, err := loadEmbeddedPages(cfg)
	if err != nil {
		logError("Failed when processing pages: %s", err)
		return err
	}
	// 设置嵌入式资源路由
	r.GET("/", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(pages))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	r.GET("/favicon.ico", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(pages))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	r.GET("/script.js", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(pages))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	r.GET("/style.css", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(pages))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	r.GET("/bootstrap.min.css", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(assets))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	r.GET("/bootstrap.bundle.min.js", func(ctx context.Context, c *app.RequestContext) {
		staticServer := http.FileServer(http.FS(assets))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			return
		}
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
	})
	return nil
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
	if cfg != nil { // 在setupLogger前添加空值检查
		setupLogger(cfg)
		setupHertZLogger(cfg)
		InitReq(cfg)
		setMemLimit(cfg)
		loadlist(cfg)
		setupRateLimit(cfg)

		if cfg.Server.Debug {
			runMode = "dev"
		} else {
			runMode = "release"
		}

		if cfg.Server.Debug {
			version = "Dev" // 如果是Debug模式，版本设置为"Dev"
		}
	}
}

func main() {
	// 如果 showVersion 为 true，则在 init 阶段已退出，这里直接返回
	if showVersion || showHelp {
		return
	}
	logDebug("Run Mode: %s", runMode)

	// 确保在程序配置加载且非版本显示模式下执行
	if cfg == nil {
		fmt.Println("Config not loaded, exiting.")
		return // 如果配置未加载，则不继续执行
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	if cfg.Server.H2C {
		r = server.New(
			server.WithHostPorts(addr),
			server.WithH2C(true),
		//	server.WithALPN(true),
		//	server.WithTransport(standard.NewTransporter),
		)
		r.AddProtocol("h2", factory.NewServerFactory())
	} else {
		r = server.New(
			server.WithHostPorts(addr),
		)
	}

	// 添加Recovery中间件
	r.Use(recovery.Recovery())
	// 添加log中间件
	r.Use(loggin.Middleware())

	setupApi(cfg, r, version)

	setupPages(cfg, r)

	/*
		// 1. GitHub Releases/Archive - Use distinct path segments for type
		r.GET("/github.com/:username/:repo/releases/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for releases
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		r.GET("/github.com/:username/:repo/archive/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for archive
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		// 2. GitHub Blob/Raw - Use distinct path segments for type
		r.GET("/github.com/:username/:repo/blob/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for blob
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		r.GET("/github.com/:username/:repo/raw/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for raw
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		r.GET("/github.com/:username/:repo/info/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for info
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})
		r.GET("/github.com/:username/:repo/git-upload-pack", func(ctx context.Context, c *app.RequestContext) {
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		// 4. Raw GitHubusercontent - Keep as is (assuming it's distinct enough)
		r.GET("/raw.githubusercontent.com/:username/:repo/*filepath", func(ctx context.Context, c *app.RequestContext) {
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		// 5. Gist GitHubusercontent - Keep as is (assuming it's distinct enough)
		r.GET("/gist.githubusercontent.com/:username/*filepath", func(ctx context.Context, c *app.RequestContext) {
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})

		// 6. GitHub API Repos - Keep as is (assuming it's distinct enough)
		r.GET("/api.github.com/repos/:username/:repo/*filepath", func(ctx context.Context, c *app.RequestContext) {
			proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
		})
	*/

	r.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter)(ctx, c)
	})

	fmt.Printf("GHProxy Version: %s\n", version)
	fmt.Printf("A Go Based High-Performance Github Proxy \n")
	fmt.Printf("Made by WJQSERVER-STUDIO\n")

	r.Spin()
	defer logger.Close()
	defer func() {
		if hertZfile != nil {
			err := hertZfile.Close()
			if err != nil {
				logError("Failed to close hertz log file: %v", err)
			}
		}
	}()
	fmt.Println("Program Exit")
}
