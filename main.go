package main

import (
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
	"ghproxy/proxy"

	"github.com/WJQSERVER-STUDIO/httpc"
	"github.com/fenthope/bauth"

	"ghproxy/weakcache"

	"github.com/fenthope/ikumi"
	"github.com/fenthope/ipfilter"
	"github.com/fenthope/reco"
	"github.com/fenthope/record"
	"github.com/infinite-iroha/touka"
	"github.com/wjqserver/modembed"
	"golang.org/x/time/rate"

	_ "net/http/pprof"
)

var (
	cfg         *config.Config
	r           *touka.Engine
	configfile  = "/data/ghproxy/config/config.toml"
	httpClient  *httpc.Client
	cfgfile     string
	version     string
	runMode     string
	showVersion bool
	showHelp    bool
)

var (
	//go:embed pages/*
	pagesFS embed.FS
)

var (
	wcache *weakcache.Cache[string] // docker token缓存
)

var (
	logger     *reco.Logger
	logDump    = logger.Debugf
	logDebug   = logger.Debugf
	logInfo    = logger.Infof
	logWarning = logger.Warnf
	logError   = logger.Errorf
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
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	recoLevel := reco.ParseLevel(cfg.Log.Level)
	logger, err = reco.New(reco.Config{
		Level:          recoLevel,
		Mode:           reco.ModeText,
		FilePath:       cfg.Log.LogFilePath,
		MaxFileSizeMB:  cfg.Log.MaxLogSize,
		EnableRotation: true,
		Async:          true,
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger.SetLevel(recoLevel)

	fmt.Printf("Log Level: %s\n", cfg.Log.Level)
	logger.Debugf("Config File Path: %s", cfgfile)
	logger.Debugf("Loaded config: %v", cfg)
	logger.Infof("Logger Initialized Successfully")
}

func setMemLimit(cfg *config.Config) {
	if cfg.Server.MemLimit > 0 {
		debug.SetMemoryLimit((cfg.Server.MemLimit) * 1024 * 1024)
		logInfo("Set Memory Limit to %d MB", cfg.Server.MemLimit)
	}
}

func loadlist(cfg *config.Config) {
	err := auth.ListInit(cfg)
	if err != nil {
		logger.Errorf("Failed to initialize list: %v", err)
	}

}

func setupApi(cfg *config.Config, r *touka.Engine, version string) {
	api.InitHandleRouter(cfg, r, version)
}

func InitReq(cfg *config.Config) {
	var err error
	httpClient, err = proxy.InitReq(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize request: %v\n", err)
		os.Exit(1)
	}
}

// loadEmbeddedPages 加载嵌入式页面资源
func loadEmbeddedPages(cfg *config.Config) (fs.FS, fs.FS, error) {
	pageFS := modembed.NewModTimeFS(pagesFS, time.Now())
	var pages fs.FS
	var err error
	switch cfg.Pages.Theme {
	case "bootstrap":
		pages, err = fs.Sub(pageFS, "pages/bootstrap")
	case "nebula":
		pages, err = fs.Sub(pageFS, "pages/nebula")
	case "design":
		pages, err = fs.Sub(pageFS, "pages/design")
	case "metro":
		pages, err = fs.Sub(pageFS, "pages/metro")
	case "classic":
		pages, err = fs.Sub(pageFS, "pages/classic")
	case "mino":
		pages, err = fs.Sub(pageFS, "pages/mino")
	case "hub":
		pages, err = fs.Sub(pageFS, "pages/hub")
	case "free":
		pages, err = fs.Sub(pageFS, "pages/free")
	default:
		pages, err = fs.Sub(pageFS, "pages/design") // 默认主题
		logWarning("Invalid Pages Theme: %s, using default theme 'design'", cfg.Pages.Theme)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to load embedded pages: %w", err)
	}

	// 初始化errPagesFs
	errPagesInitErr := proxy.InitErrPagesFS(pageFS)
	if errPagesInitErr != nil {
		logWarning("errPagesInitErr: %s", errPagesInitErr)
	}

	var assets fs.FS
	assets, err = fs.Sub(pageFS, "pages/assets")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load embedded assets: %w", err)
	}
	return pages, assets, nil
}

// setupPages 设置页面路由
func setupPages(cfg *config.Config, r *touka.Engine) {
	switch cfg.Pages.Mode {
	case "internal":
		err := setInternalRoute(cfg, r)
		if err != nil {
			logError("Failed when processing internal pages: %s", err)
			fmt.Println(err.Error())
			return
		}

	case "external":
		r.SetUnMatchFS(http.Dir(cfg.Pages.StaticDir))

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

var viaString string = "WJQSERVER-STUDIO/GHProxy"

func pageCacheHeader() func(c *touka.Context) {
	return func(c *touka.Context) {
		c.AddHeader("Cache-Control", "public, max-age=3600, must-revalidate")
		c.Next()
	}
}

func viaHeader() func(c *touka.Context) {
	return func(c *touka.Context) {
		protoVersion := fmt.Sprintf("%d.%d", c.Request.ProtoMajor, c.Request.ProtoMinor)
		c.AddHeader("Via", protoVersion+" "+viaString)
		c.Next()
	}
}

func setInternalRoute(cfg *config.Config, r *touka.Engine) error {

	// 加载嵌入式资源
	pages, assets, err := loadEmbeddedPages(cfg)
	if err != nil {
		logError("Failed when processing pages: %s", err)
		return err
	}

	r.HandleFunc([]string{"GET"}, "/favicon.ico", pageCacheHeader(), touka.FileServer(http.FS(assets)))
	r.HandleFunc([]string{"GET"}, "/", pageCacheHeader(), touka.FileServer(http.FS(pages)))
	r.HandleFunc([]string{"GET"}, "/script.js", pageCacheHeader(), touka.FileServer(http.FS(pages)))
	r.HandleFunc([]string{"GET"}, "/style.css", pageCacheHeader(), touka.FileServer(http.FS(pages)))
	r.HandleFunc([]string{"GET"}, "/bootstrap.min.css", pageCacheHeader(), touka.FileServer(http.FS(assets)))
	r.HandleFunc([]string{"GET"}, "/bootstrap.bundle.min.js", pageCacheHeader(), touka.FileServer(http.FS(assets)))

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
		InitReq(cfg)
		setMemLimit(cfg)
		loadlist(cfg)
		if cfg.Docker.Enabled {
			wcache = proxy.InitWeakCache()
		}

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
	if showVersion || showHelp {
		return
	}

	if cfg == nil {
		fmt.Println("Config not loaded, exiting.")
		return
	}

	r := touka.Default()
	r.SetProtocols(&touka.ProtocolsConfig{
		Http1:           true,
		Http2_Cleartext: true,
	})

	r.Use(touka.Recovery()) // Recovery中间件
	r.SetLogger(logger)
	r.SetErrorHandler(proxy.UnifiedToukaErrorHandler)
	r.SetHTTPClient(httpClient)
	r.Use(record.Middleware()) // log中间件
	r.Use(viaHeader())
	/*
		r.Use(compress.Compression(compress.CompressOptions{
			Algorithms: map[string]compress.AlgorithmConfig{
				compress.EncodingGzip: {
					Level:       gzip.BestCompression, // Gzip最高压缩比
					PoolEnabled: true,                 // 启用Gzip压缩器的对象池
				},
				compress.EncodingDeflate: {
					Level:       flate.DefaultCompression, // Deflate默认压缩比
					PoolEnabled: false,                    // Deflate不启用对象池
				},
				compress.EncodingZstd: {
					Level:       int(zstd.SpeedBestCompression), // Zstandard最佳压缩比
					PoolEnabled: true,                           // 启用Zstandard压缩器的对象池
				},
			},
		}))
	*/

	if cfg.RateLimit.Enabled {
		r.Use(ikumi.TokenRateLimit(ikumi.TokenRateLimiterOptions{
			Limit: rate.Limit(cfg.RateLimit.RatePerMinute),
			Burst: cfg.RateLimit.Burst,
		}))
	}

	if cfg.IPFilter.Enabled {
		var err error
		ipAllowList, ipBlockList, err := auth.ReadIPFilterList(cfg)
		if err != nil {
			fmt.Printf("Failed to read IP filter list: %v\n", err)
			logger.Errorf("Failed to read IP filter list: %v", err)
			os.Exit(1)
		}
		ipBlockFilter, err := ipfilter.NewIPFilter(ipfilter.IPFilterConfig{
			EnableAllowList: cfg.IPFilter.EnableAllowList,
			EnableBlockList: cfg.IPFilter.EnableBlockList,
			AllowList:       ipAllowList,
			BlockList:       ipBlockList,
		})
		if err != nil {
			fmt.Printf("Failed to initialize IP filter: %v\n", err)
			logger.Errorf("Failed to initialize IP filter: %v", err)
			os.Exit(1)
		} else {
			r.Use(ipBlockFilter)
		}
	}
	setupApi(cfg, r, version)
	setupPages(cfg, r)
	r.SetRedirectTrailingSlash(false)

	r.GET("/github.com/:user/:repo/releases/*filepath", func(c *touka.Context) {
		c.Set("matcher", "releases")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/github.com/:user/:repo/archive/*filepath", func(c *touka.Context) {
		c.Set("matcher", "releases")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/github.com/:user/:repo/blob/*filepath", func(c *touka.Context) {
		c.Set("matcher", "blob")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/github.com/:user/:repo/raw/*filepath", func(c *touka.Context) {
		c.Set("matcher", "raw")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/github.com/:user/:repo/info/*filepath", func(c *touka.Context) {
		c.Set("matcher", "clone")
		proxy.RoutingHandler(cfg)(c)
	})
	r.GET("/github.com/:user/:repo/git-upload-pack", func(c *touka.Context) {
		c.Set("matcher", "clone")
		proxy.RoutingHandler(cfg)(c)
	})
	r.POST("/github.com/:user/:repo/git-upload-pack", func(c *touka.Context) {
		c.Set("matcher", "clone")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/raw.githubusercontent.com/:user/:repo/*filepath", func(c *touka.Context) {
		c.Set("matcher", "raw")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/gist.githubusercontent.com/:user/*filepath", func(c *touka.Context) {
		c.Set("matcher", "gist")
		proxy.NoRouteHandler(cfg)(c)
	})

	r.ANY("/api.github.com/repos/:user/:repo/*filepath", func(c *touka.Context) {
		c.Set("matcher", "api")
		proxy.RoutingHandler(cfg)(c)
	})

	r.GET("/v2/",
		r.UseIf(cfg.Docker.Auth, func() touka.HandlerFunc {
			return bauth.BasicAuthForStatic(cfg.Docker.Credentials, "GHProxy Docker Proxy")
		}),
		func(c *touka.Context) {
			emptyJSON := "{}"
			c.Header("Content-Type", "application/json")
			c.Header("Content-Length", fmt.Sprint(len(emptyJSON)))

			c.Header("Docker-Distribution-API-Version", "registry/2.0")

			c.Status(200)
			c.Writer.Write([]byte(emptyJSON))
		},
	)

	r.GET("/v2", func(c *touka.Context) {
		// 重定向到 /v2/
		c.Redirect(http.StatusMovedPermanently, "/v2/")
	})

	r.ANY("/v2/:target/:user/:repo/*filepath", func(c *touka.Context) {
		proxy.GhcrWithImageRouting(cfg)(c)
	})

	r.NoRoute(func(c *touka.Context) {
		proxy.NoRouteHandler(cfg)(c)
	})

	fmt.Printf("GHProxy Version: %s\n", version)
	fmt.Printf("A Go Based High-Performance Github Proxy \n")
	fmt.Printf("Made by WJQSERVER-STUDIO\n")
	fmt.Printf("Power by Touka\n")

	if cfg.Server.Debug {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	if wcache != nil {
		defer wcache.StopCleanup()
	}

	defer logger.Close()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	err := r.RunShutdown(addr)
	if err != nil {
		logError("Server Run Error: %v", err)
		fmt.Printf("Server Run Error: %v\n", err)
	}

	fmt.Println("Program Exit")
}
