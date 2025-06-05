package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/rate"

	"github.com/cloudwego/hertz/pkg/app"
)

func listCheck(cfg *config.Config, c *app.RequestContext, user string, repo string, rawPath string) bool {
	if cfg.Auth.ForceAllowApi && cfg.Auth.ForceAllowApiPassList {
		return false
	}
	// 白名单检查
	if cfg.Whitelist.Enabled {
		whitelist := auth.CheckWhitelist(user, repo)
		if !whitelist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Whitelist Blocked repo: %s/%s", user, repo)))
			logInfo("%s %s %s %s %s Whitelist Blocked repo: %s/%s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
			return true
		}
	}

	// 黑名单检查
	if cfg.Blacklist.Enabled {
		blacklist := auth.CheckBlacklist(user, repo)
		if blacklist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Blacklist Blocked repo: %s/%s", user, repo)))
			logInfo("%s %s %s %s %s Blacklist Blocked repo: %s/%s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
			return true
		}
	}

	return false
}

// 鉴权
func authCheck(c *app.RequestContext, cfg *config.Config, matcher string, rawPath string) bool {
	var err error

	if matcher == "api" && !cfg.Auth.ForceAllowApi {
		if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
			ErrorPage(c, NewErrorWithStatusLookup(403, "Github API Req without AuthHeader is Not Allowed"))
			logInfo("%s %s %s AuthHeader Unavailable", c.ClientIP(), c.Method(), rawPath)
			return true
		}
	}

	// 鉴权
	if cfg.Auth.Enabled {
		var authcheck bool
		authcheck, err = auth.AuthHandler(c, cfg)
		if !authcheck {
			ErrorPage(c, NewErrorWithStatusLookup(401, fmt.Sprintf("Unauthorized: %v", err)))
			logInfo("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), err)
			return true
		}
	}

	return false
}

func rateCheck(cfg *config.Config, c *app.RequestContext, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter) bool {
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
			ErrorPage(c, NewErrorWithStatusLookup(500, "Invalid RateLimit Method"))
			return true
		}

		if !allowed {
			ErrorPage(c, NewErrorWithStatusLookup(429, fmt.Sprintf("Too Many Requests; Rate Limit is %d per minute", cfg.RateLimit.RatePerMinute)))
			logInfo("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Method(), c.Request.RequestURI(), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
			return true
		}
	}

	return false
}
