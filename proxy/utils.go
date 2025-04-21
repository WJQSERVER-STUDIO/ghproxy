package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/rate"
	"io/fs"

	"github.com/cloudwego/hertz/pkg/app"
)

func listCheck(cfg *config.Config, c *app.RequestContext, user string, repo string, rawPath string) {
	var errMsg string

	// 白名单检查
	if cfg.Whitelist.Enabled {
		var whitelist bool
		whitelist = auth.CheckWhitelist(user, repo)
		if !whitelist {
			errMsg = fmt.Sprintf("Whitelist Blocked repo: %s/%s", user, repo)
			c.JSON(403, map[string]string{"error": errMsg})
			logWarning("%s %s %s %s %s Whitelist Blocked repo: %s/%s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
			return
		}
	}

	// 黑名单检查
	if cfg.Blacklist.Enabled {
		var blacklist bool
		blacklist = auth.CheckBlacklist(user, repo)
		if blacklist {
			errMsg = fmt.Sprintf("Blacklist Blocked repo: %s/%s", user, repo)
			c.JSON(403, map[string]string{"error": errMsg})
			logWarning("%s %s %s %s %s Blacklist Blocked repo: %s/%s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
			return
		}
	}
}

// 鉴权
func authCheck(c *app.RequestContext, cfg *config.Config, matcher string, rawPath string) {
	var err error

	if matcher == "api" && !cfg.Auth.ForceAllowApi {
		if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
			c.JSON(403, map[string]string{"error": "Github API Req without AuthHeader is Not Allowed"})
			logWarning("%s %s %s %s %s AuthHeader Unavailable", c.ClientIP(), c.Method(), rawPath)
			return
		}
	}

	// 鉴权
	if cfg.Auth.Enabled {
		var authcheck bool
		authcheck, err = auth.AuthHandler(c, cfg)
		if !authcheck {
			c.JSON(401, map[string]string{"error": "Unauthorized"})
			logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), err)
			return
		}
	}
}

func rateCheck(cfg *config.Config, c *app.RequestContext, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter) {
	// 限制访问频率
	if cfg.RateLimit.Enabled {

		var allowed bool

		switch cfg.RateLimit.RateMethod {
		case "ip":
			allowed = iplimiter.Allow(c.ClientIP())
		case "total":
			allowed = limiter.Allow()
		default:
			logWarning("Invalid RateLimit Method")
			c.JSON(500, map[string]string{"error": "Invalid RateLimit Method"})
			return
		}

		if !allowed {
			c.JSON(429, map[string]string{"error": "Too Many Requests"})
			logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Method(), c.Request.RequestURI(), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
			return
		}
	}
}

var errPagesFs fs.FS

func InitErrPagesFS(pages fs.FS) error {
	var err error
	errPagesFs, err = fs.Sub(pages, "pages/err")
	if err != nil {
		return err
	}
	return nil
}

func NotFoundPage(c *app.RequestContext) {
	pageData, err := fs.ReadFile(errPagesFs, "404.html")
	if err != nil {
		c.JSON(404, map[string]string{"error": "Not Found"})
		logDebug("Error reading 404.html: %v", err)
		return
	}
	c.Data(404, "text/html; charset=utf-8", pageData)
	return
}
