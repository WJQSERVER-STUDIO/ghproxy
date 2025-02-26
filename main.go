package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"ghproxy/api"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/loggin"
	"ghproxy/proxy"
	"ghproxy/rate"
	"ghproxy/timing"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	cfg        *config.Config
	router     *gin.Engine
	configfile = "/data/ghproxy/config/config.toml"
	cfgfile    string
	version    string
	dev        string
	runMode    string
	limiter    *rate.RateLimiter
	iplimiter  *rate.IPRateLimiter
)

var (
	//go:embed pages/bootstrap/*
	pagesFS embed.FS
	//go:embed pages/nebula/*
	NebulaPagesFS embed.FS
)

var (
	logw       = logger.Logw
	LogDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func readFlag() {
	flag.StringVar(&cfgfile, "cfg", configfile, "config file path")
}

func loadConfig() {
	var err error
	cfg, err = config.LoadConfig(cfgfile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
	}
	if cfg.Server.Debug {
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
func loadEmbeddedPages(cfg *config.Config) (fs.FS, error) {
	var pages fs.FS
	var err error

	switch cfg.Pages.Theme {
	case "bootstrap":
		pages, err = fs.Sub(pagesFS, "pages/bootstrap")
	case "nebula":
		pages, err = fs.Sub(NebulaPagesFS, "pages/nebula")
	default:
		pages, err = fs.Sub(pagesFS, "pages/bootstrap") // 默认主题
		logWarning("Invalid Pages Theme: %s, using default theme 'bootstrap'", cfg.Pages.Theme)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load embedded pages: %w", err)
	}
	return pages, nil
}

// setupPages 设置页面路由
func setupPages(cfg *config.Config, router *gin.Engine) {
	switch cfg.Pages.Mode {
	case "internal":
		// 加载嵌入式资源
		pages, err := loadEmbeddedPages(cfg)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			return
		}

		// 设置嵌入式资源路由
		router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))

	case "external":
		// 设置外部资源路径
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
		stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)

		// 设置外部资源路由
		router.GET("/", func(c *gin.Context) {
			c.File(indexPagePath)
			logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
		router.StaticFile("/favicon.ico", faviconPath)
		router.StaticFile("/script.js", javascriptsPath)
		router.StaticFile("/style.css", stylesheetsPath)

	default:
		// 处理无效的Pages Mode
		logWarning("Invalid Pages Mode: %s, using default embedded theme", cfg.Pages.Mode)

		// 加载嵌入式资源
		pages, err := loadEmbeddedPages(cfg)
		if err != nil {
			logError("Failed when processing pages: %s", err)
			return
		}
		// 设置嵌入式资源路由
		router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))
	}
}

func init() {
	readFlag()
	flag.Parse()
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

	//H2C默认值为true，而后遵循cfg.Server.EnableH2C的设置
	if cfg.Server.H2C {
		router.UseH2C = true
	} else {
		logWarning("cfg.Server.EnableH2C 将于2.4.0弃用，请使用cfg.Server.H2C")
		if cfg.Server.EnableH2C == "on" {
			router.UseH2C = true
		} else if cfg.Server.EnableH2C == "" {
			router.UseH2C = true
		} else {
			router.UseH2C = false
		}
	}

	/*
		// (2.4.0启用)
		if cfg.Server.H2C {
			router.UseH2C = true
		}
	*/

	setupApi(cfg, router, version)

	// setupPages(cfg, router)  // 2.4.0启用

	logInfo("Pages Mode: %s", cfg.Pages.Mode)
	if cfg.Pages.Mode == "internal" {
		var pages fs.FS
		var err error
		if cfg.Pages.Theme == "bootstrap" {
			pages, err = fs.Sub(pagesFS, "pages/bootstrap")
			if err != nil {
				logError("Failed when processing pages: %s", err)
			}
		} else if cfg.Pages.Theme == "nebula" {
			pages, err = fs.Sub(NebulaPagesFS, "pages/nebula")
			if err != nil {
				logError("Failed when processing pages: %s", err)
			}
		} else {
			pages, err = fs.Sub(pagesFS, "pages/bootstrap")
			if err != nil {
				logError("Failed when processing pages: %s", err)
			}
		}
		router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
		router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))
	} else if cfg.Pages.Mode == "external" {
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
		stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)
		router.GET("/", func(c *gin.Context) {
			c.File(indexPagePath)
			logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
		router.StaticFile("/favicon.ico", faviconPath)
		router.StaticFile("/script.js", javascriptsPath)
		router.StaticFile("/style.css", stylesheetsPath)
	} else {
		logWarning("缺少 cfg.Pages.Mode 配置, cfg.Pages.Enable 将于2.4.0弃用，请使用cfg.Pages.Mode")
		if cfg.Pages.Enabled {
			indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
			faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
			javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
			stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)
			router.GET("/", func(c *gin.Context) {
				c.File(indexPagePath)
				logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
			})
			router.StaticFile("/favicon.ico", faviconPath)
			router.StaticFile("/script.js", javascriptsPath)
			router.StaticFile("/style.css", stylesheetsPath)
		} else if !cfg.Pages.Enabled {
			var pages fs.FS
			var err error
			if cfg.Pages.Theme == "bootstrap" {
				pages, err = fs.Sub(pagesFS, "pages/bootstrap")
				if err != nil {
					logError("Failed when processing pages: %s", err)
				}
			} else if cfg.Pages.Theme == "nebula" {
				pages, err = fs.Sub(NebulaPagesFS, "pages/nebula")
				if err != nil {
					logError("Failed when processing pages: %s", err)
				}
			} else {
				pages, err = fs.Sub(pagesFS, "pages/bootstrap")
				if err != nil {
					logError("Failed when processing pages: %s", err)
				}
			}
			router.GET("/", gin.WrapH(http.FileServer(http.FS(pages))))
			router.GET("/favicon.ico", gin.WrapH(http.FileServer(http.FS(pages))))
			router.GET("/script.js", gin.WrapH(http.FileServer(http.FS(pages))))
			router.GET("/style.css", gin.WrapH(http.FileServer(http.FS(pages))))
		}
	}

	router.NoRoute(func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(c)
	})

	fmt.Printf("GHProxy Version: %s\n", version)
	fmt.Printf("A Go Based High-Performance Github Proxy \n")
	fmt.Printf("Made by WJQSERVER-STUDIO\n")
}

func main() {
	err := router.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logError("Failed to start server: %v\n", err)
	}
	defer logger.Close()
	fmt.Println("Program Exit")
}
