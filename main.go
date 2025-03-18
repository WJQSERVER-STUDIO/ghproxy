package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
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

	"github.com/hertz-contrib/http2/factory"
)

var (
	cfg        *config.Config
	r          *server.Hertz
	configfile = "/data/ghproxy/config/config.toml"
	cfgfile    string
	version    string
	runMode    string
	limiter    *rate.RateLimiter
	iplimiter  *rate.IPRateLimiter
)

var (
	//go:embed pages/*
	pagesFS embed.FS
	/*
		//go:embed pages/bootstrap/*
		BootstrapPagesFS embed.FS
		//go:embed pages/nebula/*
		NebulaPagesFS embed.FS
		//go:embed pages/design/*
		DesignPagesFS embed.FS
	*/
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
func loadEmbeddedPages(cfg *config.Config) (fs.FS, error) {
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
func setupPages(cfg *config.Config, r *server.Hertz) {
	switch cfg.Pages.Mode {
	case "internal":
		// 加载嵌入式资源
		pages, err := loadEmbeddedPages(cfg)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			return
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

	case "external":
		// 设置外部资源路径
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		javascriptsPath := fmt.Sprintf("%s/script.js", cfg.Pages.StaticDir)
		stylesheetsPath := fmt.Sprintf("%s/style.css", cfg.Pages.StaticDir)
		//bootstrapPath := fmt.Sprintf("%s/bootstrap.min.css", cfg.Pages.StaticDir)

		// 设置外部资源路由
		r.StaticFile("/", indexPagePath)
		r.StaticFile("/favicon.ico", faviconPath)
		r.StaticFile("/script.js", javascriptsPath)
		r.StaticFile("/style.css", stylesheetsPath)
		//router.StaticFile("/bootstrap.min.css", bootstrapPath)

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
		runMode = "dev"
	} else {
		runMode = "release"
	}

	if cfg.Server.Debug {
		version = "Dev"
	}

}

func main() {
	logDebug("Run Mode: %s", runMode)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	r := server.New(
		server.WithHostPorts(addr),
		server.WithH2C(true),
	)

	r.AddProtocol("h2", factory.NewServerFactory())

	// 添加Recovery中间件
	r.Use(recovery.Recovery())
	// 添加log中间件
	r.Use(loggin.Middleware())

	setupApi(cfg, r, version)

	setupPages(cfg, r)

	// 1. GitHub Releases/Archive - Use distinct path segments for type
	r.GET("/github.com/:username/:repo/releases/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for releases
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	r.GET("/github.com/:username/:repo/archive/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for archive
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	// 2. GitHub Blob/Raw - Use distinct path segments for type
	r.GET("/github.com/:username/:repo/blob/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for blob
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	r.GET("/github.com/:username/:repo/raw/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for raw
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	r.GET("/github.com/:username/:repo/info/*filepath", func(ctx context.Context, c *app.RequestContext) { // Distinct path for info
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})
	r.GET("/github.com/:username/:repo/git-upload-pack", func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	// 4. Raw GitHubusercontent - Keep as is (assuming it's distinct enough)
	r.GET("/raw.githubusercontent.com/:username/:repo/*filepath", func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	// 5. Gist GitHubusercontent - Keep as is (assuming it's distinct enough)
	r.GET("/gist.githubusercontent.com/:username/*filepath", func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	// 6. GitHub API Repos - Keep as is (assuming it's distinct enough)
	r.GET("/api.github.com/repos/:username/:repo/*filepath", func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	r.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter, runMode)(ctx, c)
	})

	fmt.Printf("GHProxy Version: %s\n", version)
	fmt.Printf("A Go Based High-Performance Github Proxy \n")
	fmt.Printf("Made by WJQSERVER-STUDIO\n")

	r.Spin()
	defer logger.Close()
	fmt.Println("Program Exit")
}
