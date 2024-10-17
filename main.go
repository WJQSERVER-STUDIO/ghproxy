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
	LogWarning = logger.LogWarning
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
	// 注册 API 接口
	api.InitHandleRouter(cfg, router)
}

func init() {
	readFlag()
	flag.Parse()
	loadConfig()
	setupLogger(cfg)
	loadlist(cfg)

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化路由
	router = gin.Default()

	setupApi(cfg, router)

	// 定义路由
	router.GET("/", func(c *gin.Context) {
		// 返回403错误
		c.String(http.StatusForbidden, "403 Forbidden This route is not allowed to access.")
		// 记录访问者IP UA METHOD
		LogWarning("Forbidden: IP:%s UA:%s METHOD:%s", c.ClientIP(), c.Request.UserAgent(), c.Request.Method)
	})

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
