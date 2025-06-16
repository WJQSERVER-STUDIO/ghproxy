package api

import (
	"ghproxy/config"
	"ghproxy/middleware/nocache"

	"github.com/infinite-iroha/touka"
)

func InitHandleRouter(cfg *config.Config, r *touka.Engine, version string) {
	apiRouter := r.Group("/api", nocache.NoCacheMiddleware())
	{
		apiRouter.GET("/size_limit", func(c *touka.Context) {
			SizeLimitHandler(cfg, c)
		})
		apiRouter.GET("/whitelist/status", func(c *touka.Context) {
			WhiteListStatusHandler(cfg, c)
		})
		apiRouter.GET("/blacklist/status", func(c *touka.Context) {
			BlackListStatusHandler(cfg, c)
		})
		apiRouter.GET("/cors/status", func(c *touka.Context) {
			CorsStatusHandler(cfg, c)
		})
		apiRouter.GET("/healthcheck", func(c *touka.Context) {
			HealthcheckHandler(c)
		})
		apiRouter.GET("/ok", func(c *touka.Context) {
			HealthcheckHandler(c)
		})
		apiRouter.GET("/version", func(c *touka.Context) {
			VersionHandler(c, version)
		})
		apiRouter.GET("/rate_limit/status", func(c *touka.Context) {
			RateLimitStatusHandler(cfg, c)
		})
		apiRouter.GET("/rate_limit/limit", func(c *touka.Context) {
			RateLimitLimitHandler(cfg, c)
		})
		apiRouter.GET("/smartgit/status", func(c *touka.Context) {
			SmartGitStatusHandler(cfg, c)
		})
		apiRouter.GET("/shell_nest/status", func(c *touka.Context) {
			shellNestStatusHandler(cfg, c)
		})
		apiRouter.GET("/oci_proxy/status", func(c *touka.Context) {
			ociProxyStatusHandler(cfg, c)
		})
	}
}

func SizeLimitHandler(cfg *config.Config, c *touka.Context) {
	sizeLimit := cfg.Server.SizeLimit
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"MaxResponseBodySize": sizeLimit,
	}))
}

func WhiteListStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Whitelist": cfg.Whitelist.Enabled,
	}))
}

func BlackListStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Blacklist": cfg.Blacklist.Enabled,
	}))
}

func CorsStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Cors": cfg.Server.Cors,
	}))
}

func HealthcheckHandler(c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Status": "OK",
		"Repo":   "WJQSERVER-STUDIO/GHProxy",
		"Author": "WJQSERVER-STUDIO",
	}))
}

func VersionHandler(c *touka.Context, version string) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"Version": version,
		"Repo":    "WJQSERVER-STUDIO/GHProxy",
		"Author":  "WJQSERVER-STUDIO",
	}))
}

func RateLimitStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"RateLimit": cfg.RateLimit.Enabled,
	}))
}

func RateLimitLimitHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"RatePerMinute": cfg.RateLimit.RatePerMinute,
	}))
}

func SmartGitStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.GitClone.Mode == "cache",
	}))
}

func shellNestStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.Shell.Editor,
	}))
}

func ociProxyStatusHandler(cfg *config.Config, c *touka.Context) {
	c.SetHeader("Content-Type", "application/json")
	c.JSON(200, (map[string]interface{}{
		"enabled": cfg.Docker.Enabled,
		"target":  cfg.Docker.Target,
	}))
}
