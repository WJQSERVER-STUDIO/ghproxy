package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"ghproxy/api"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/logger"
	"ghproxy/proxy"
	"ghproxy/rate"

	"github.com/gin-gonic/gin"
)

var (
	cfg        *config.Config
	router     *gin.Engine
	configfile = "/data/ghproxy/config/config.toml"
	cfgfile    string
	version    string
	limiter    *rate.RateLimiter
	iplimiter  *rate.IPRateLimiter
)

var (
	logw       = logger.Logw
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
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Config File Path: ", cfgfile)
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger(cfg *config.Config) {
	var err error
	err = logger.Init(cfg.Log.LogFilePath, cfg.Log.MaxLogSize)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logInfo("Config File Path: ", cfgfile)
	logInfo("Loaded config: %v\n", cfg)
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
		logInfo("Rate Limit Loaded")
	}
}

func init() {
	readFlag()
	flag.Parse()
	loadConfig()
	setupLogger(cfg)
	loadlist(cfg)
	setupRateLimit(cfg)

	gin.SetMode(gin.ReleaseMode)

	router = gin.Default()
	//H2C默认值为true，而后遵循cfg.Server.EnableH2C的设置
	if cfg.Server.EnableH2C == "on" {
		router.UseH2C = true
	} else if cfg.Server.EnableH2C == "" {
		router.UseH2C = true
	} else {
		router.UseH2C = false
	}
	/*if !cfg.Server.EnableH2C {
		router.UseH2C = false
	} else {
		router.UseH2C = true
	}*/

	setupApi(cfg, router, version)

	if cfg.Pages.Enabled {
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		router.GET("/", func(c *gin.Context) {
			c.File(indexPagePath)
			logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
		router.StaticFile("/favicon.ico", faviconPath)
	} else if !cfg.Pages.Enabled {
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusForbidden, "403 Forbidden Access")
			logWarning("403 > Path:/ IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
	}

	router.NoRoute(func(c *gin.Context) {
		proxy.NoRouteHandler(cfg, limiter, iplimiter)(c)
	})
}

func main() {
	err := router.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logError("Failed to start server: %v\n", err)
	}
	defer logger.Close()
	fmt.Println("Program Exit")
}
