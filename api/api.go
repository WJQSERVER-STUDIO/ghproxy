package api

import (
	"encoding/json"
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	cfg    *config.Config
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	LogWarning = logger.LogWarning
	logError   = logger.LogError
)

func InitHandleRouter(cfg *config.Config, router *gin.Engine) {
	// 设置路由
	apiRouter := router.Group("api")
	{
		apiRouter.GET("/size_limit", func(c *gin.Context) {
			SizeLimitHandler(cfg, c)
		})
		apiRouter.GET("/whitelist/status", func(c *gin.Context) {
			WhiteListStatusHandler(c, cfg)
		})
		apiRouter.GET("/blacklist/status", func(c *gin.Context) {
			BlackListStatusHandler(c, cfg)
		})
		apiRouter.GET("/healthcheck", func(c *gin.Context) {
			HealthcheckHandler(c)
		})
	}
	logInfo("API router Init success")
}

func SizeLimitHandler(cfg *config.Config, c *gin.Context) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"MaxResponseBodySize": cfg.Server.SizeLimit,
	})
}

func WhiteListStatusHandler(c *gin.Context, cfg *config.Config) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Whitelist": cfg.Whitelist.Enabled,
	})
}

func BlackListStatusHandler(c *gin.Context, cfg *config.Config) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Blacklist": cfg.Blacklist.Enabled,
	})
}

func HealthcheckHandler(c *gin.Context) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Status": "OK",
	})
}
