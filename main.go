package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"ghproxy/api"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/logger"
	"ghproxy/proxy"

	"github.com/gin-gonic/gin"
)

var (
	cfg        *config.Config
	router     *gin.Engine
	configfile = "/data/ghproxy/config/config.toml"
	cfgfile    string
)

// 日志模块
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
	// 初始化配置
	cfg, err = config.LoadConfig(cfgfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Config File Path: ", cfgfile)
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger(cfg *config.Config) {
	// 初始化日志模块
	var err error
	err = logger.Init(cfg.Log.LogFilePath, cfg.Log.MaxLogSize) // 传递日志文件路径
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

func setupApi(cfg *config.Config, router *gin.Engine) {
	api.InitHandleRouter(cfg, router)
}

func init() {
	readFlag()
	flag.Parse()
	loadConfig()
	setupLogger(cfg)
	loadlist(cfg)

	gin.SetMode(gin.ReleaseMode)

	router = gin.Default()
	router.UseH2C = true

	setupApi(cfg, router)

	if cfg.Pages.Enabled {
		indexPagePath := fmt.Sprintf("%s/index.html", cfg.Pages.StaticDir)
		faviconPath := fmt.Sprintf("%s/favicon.ico", cfg.Pages.StaticDir)
		// 静态index页
		//router.StaticFile("/", indexPagePath)
		router.GET("/", func(c *gin.Context) {
			c.File(indexPagePath)
			logInfo("IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
		// 静态favicon.ico
		router.StaticFile("/favicon.ico", faviconPath)
	} else if !cfg.Pages.Enabled {
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusForbidden, "403 Forbidden This route is not allowed to access.")
			logWarning("Forbidden: IP:%s UA:%s METHOD:%s HTTPv:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.Proto)
		})
	}

	// 未匹配路由处理
	router.NoRoute(func(c *gin.Context) {
		proxy.NoRouteHandler(cfg)(c)
	})
}

func main() {
	// 启动服务器
	err := router.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logError("Error starting server: %v\n", err)
	}

	fmt.Println("Program finished")
}
