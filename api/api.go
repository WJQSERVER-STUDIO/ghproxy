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

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func InitHandleRouter(cfg *config.Config, router *gin.Engine, version string) {
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
		apiRouter.GET("/cors/status", func(c *gin.Context) {
			CorsStatusHandler(c, cfg)
		})
		apiRouter.GET("/healthcheck", func(c *gin.Context) {
			HealthcheckHandler(c)
		})
		apiRouter.GET("/version", func(c *gin.Context) {
			VersionHandler(c, version)
		})
		apiRouter.GET("/rate_limit/status", func(c *gin.Context) {
			RateLimitStatusHandler(c, cfg)
		})
	}
	logInfo("API router Init success")
}

func SizeLimitHandler(cfg *config.Config, c *gin.Context) {
	sizeLimit := cfg.Server.SizeLimit
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"MaxResponseBodySize": sizeLimit,
	})
}

func WhiteListStatusHandler(c *gin.Context, cfg *config.Config) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Whitelist": cfg.Whitelist.Enabled,
	})
}

func BlackListStatusHandler(c *gin.Context, cfg *config.Config) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Blacklist": cfg.Blacklist.Enabled,
	})
}

func CorsStatusHandler(c *gin.Context, cfg *config.Config) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Cors": cfg.CORS.Enabled,
	})
}

func HealthcheckHandler(c *gin.Context) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Status": "OK",
	})
}

func VersionHandler(c *gin.Context, version string) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"Version": version,
	})
}

func RateLimitStatusHandler(c *gin.Context, cfg *config.Config) {
	logInfo("%s %s %s %s %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto)
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"RateLimit": cfg.RateLimit.Enabled,
	})
}
