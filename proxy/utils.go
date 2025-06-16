package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

func listCheck(cfg *config.Config, c *touka.Context, user string, repo string, rawPath string) bool {
	if cfg.Auth.ForceAllowApi && cfg.Auth.ForceAllowApiPassList {
		return false
	}
	// 白名单检查
	if cfg.Whitelist.Enabled {
		whitelist := auth.CheckWhitelist(user, repo)
		if !whitelist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Whitelist Blocked repo: %s/%s", user, repo)))
			c.Infof("%s %s %s %s %s Whitelist Blocked repo: %s/%s", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, user, repo)
			return true
		}
	}

	// 黑名单检查
	if cfg.Blacklist.Enabled {
		blacklist := auth.CheckBlacklist(user, repo)
		if blacklist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Blacklist Blocked repo: %s/%s", user, repo)))
			c.Infof("%s %s %s %s %s Blacklist Blocked repo: %s/%s", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, user, repo)
			return true
		}
	}

	return false
}

// 鉴权
func authCheck(c *touka.Context, cfg *config.Config, matcher string, rawPath string) bool {
	var err error

	if matcher == "api" && !cfg.Auth.ForceAllowApi {
		if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
			ErrorPage(c, NewErrorWithStatusLookup(403, "Github API Req without AuthHeader is Not Allowed"))
			c.Infof("%s %s %s AuthHeader Unavailable", c.ClientIP(), c.Request.Method, rawPath)
			return true
		}
	}

	// 鉴权
	if cfg.Auth.Enabled {
		var authcheck bool
		authcheck, err = auth.AuthHandler(c, cfg)
		if !authcheck {
			ErrorPage(c, NewErrorWithStatusLookup(401, fmt.Sprintf("Unauthorized: %v", err)))
			c.Infof("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, err)
			return true
		}
	}

	return false
}
