package api

import (
	"context"
	"ghproxy/config"
	"ghproxy/middleware/nocache"

	"github.com/WJQSERVER-STUDIO/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func InitHandleRouter(cfg *config.Config, r *server.Hertz, version string) {
	apiRouter := r.Group("/api", nocache.NoCacheMiddleware())
	{
		apiRouter.GET("/size_limit", func(ctx context.Context, c *app.RequestContext) {
			SizeLimitHandler(cfg, c, ctx)
		})
		apiRouter.GET("/whitelist/status", func(ctx context.Context, c *app.RequestContext) {
			WhiteListStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/blacklist/status", func(ctx context.Context, c *app.RequestContext) {
			BlackListStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/cors/status", func(ctx context.Context, c *app.RequestContext) {
			CorsStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/healthcheck", func(ctx context.Context, c *app.RequestContext) {
			HealthcheckHandler(c, ctx)
		})
		apiRouter.GET("/version", func(ctx context.Context, c *app.RequestContext) {
			VersionHandler(c, ctx, version)
		})
		apiRouter.GET("/rate_limit/status", func(ctx context.Context, c *app.RequestContext) {
			RateLimitStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/rate_limit/limit", func(ctx context.Context, c *app.RequestContext) {
			RateLimitLimitHandler(cfg, c, ctx)
		})
		apiRouter.GET("/smartgit/status", func(ctx context.Context, c *app.RequestContext) {
			SmartGitStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/shell_nest/status", func(ctx context.Context, c *app.RequestContext) {
			shellNestStatusHandler(cfg, c, ctx)
		})
		apiRouter.GET("/oci_proxy/status", func(ctx context.Context, c *app.RequestContext) {
			ociProxyStatusHandler(cfg, c, ctx)
		})
	}
	logInfo("API router Init success")
}

func SizeLimitHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	sizeLimit := cfg.Server.SizeLimit
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"MaxResponseBodySize": sizeLimit,
	}))
}

func WhiteListStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Whitelist": cfg.Whitelist.Enabled,
	}))
}

func BlackListStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Blacklist": cfg.Blacklist.Enabled,
	}))
}

func CorsStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Cors": cfg.Server.Cors,
	}))
}

func HealthcheckHandler(c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Status": "OK",
	}))
}

func VersionHandler(c *app.RequestContext, ctx context.Context, version string) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Version": version,
	}))
}

func RateLimitStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"RateLimit": cfg.RateLimit.Enabled,
	}))
}

func RateLimitLimitHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"RatePerMinute": cfg.RateLimit.RatePerMinute,
	}))
}

func SmartGitStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.GitClone.Mode == "cache",
	}))
}

func shellNestStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.Shell.Editor,
	}))
}

func ociProxyStatusHandler(cfg *config.Config, c *app.RequestContext, ctx context.Context) {
	c.Response.Header.Set("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.Docker.Enabled,
		"target":  cfg.Docker.Target,
	}))
}
