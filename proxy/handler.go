package proxy

import (
	"context"
	"errors"
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/rate"
	"net/http"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

var re = regexp.MustCompile(`^(http:|https:)?/?/?(.*)`) // 匹配http://或https://开头的路径

func NoRouteHandler(cfg *config.Config, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

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
				return
			}

			if !allowed {
				c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too Many Requests"})
				logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Method(), c.Request.RequestURI(), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
				return
			}
		}

		var (
			rawPath string
			matches []string
			errMsg  string
		)

		rawPath = strings.TrimPrefix(string(c.Request.RequestURI()), "/") // 去掉前缀/
		matches = re.FindStringSubmatch(rawPath)                          // 匹配路径
		logInfo("URL: %v", matches)

		// 匹配路径错误处理
		if len(matches) < 3 {
			errMsg = fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
			logWarning(errMsg)
			c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)
			return
		}

		// 制作url
		rawPath = "https://" + matches[2]

		var (
			user    string
			repo    string
			matcher string
			err     error
		)

		user, repo, matcher, err = Matcher(rawPath, cfg)
		if err != nil {
			if errors.Is(err, ErrInvalidURL) {
				c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)
				logWarning(err.Error())
				return
			}
			if errors.Is(err, ErrAuthHeaderUnavailable) {
				c.String(http.StatusForbidden, "AuthHeader Unavailable")
				logWarning(err.Error())
				return
			}
		}

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
		// dump log 记录详细信息 c.ClientIP(), c.Method(), rawPath,c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), full Header
		logDump("%s %s %s %s %s %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), c.Request.Header.Header())
		var repouser string
		repouser = fmt.Sprintf("%s/%s", user, repo)

		// 白名单检查
		if cfg.Whitelist.Enabled {
			var whitelist bool
			whitelist = auth.CheckWhitelist(user, repo)
			if !whitelist {
				errMsg = fmt.Sprintf("Whitelist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, map[string]string{"error": errMsg})
				logWarning("%s %s %s %s %s Whitelist Blocked repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), repouser)
				return
			}
		}

		// 黑名单检查
		if cfg.Blacklist.Enabled {
			var blacklist bool
			blacklist = auth.CheckBlacklist(user, repo)
			if blacklist {
				errMsg = fmt.Sprintf("Blacklist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, map[string]string{"error": errMsg})
				logWarning("%s %s %s %s %s Blacklist Blocked repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), repouser)
				return
			}
		}

		// 若匹配api.github.com/repos/用户名/仓库名/路径, 则检查是否开启HeaderAuth

		// 处理blob/raw路径
		if matcher == "blob" {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		// 鉴权
		if cfg.Auth.Enabled {
			var authcheck bool
			authcheck, err = auth.AuthHandler(ctx, c, cfg)
			if !authcheck {
				//c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
				c.AbortWithStatusJSON(401, map[string]string{"error": "Unauthorized"})
				logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), err)
				return
			}
		}

		// IP METHOD URL USERAGENT PROTO MATCHES
		logDebug("%s %s %s %s %s Matched: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), matcher)

		switch matcher {
		case "releases", "blob", "raw", "gist", "api":
			ChunkedProxyRequest(ctx, c, rawPath, cfg, matcher)
		case "clone":
			GitReq(ctx, c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			fmt.Println("Invalid input.")
			return
		}
	}
}

func RoutingHandler(cfg *config.Config, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 输出所有传入参数
		logDebug("All Request Params: %v", c.Params)
		logDebug("Context Params(Matcher): %v", ctx.Value("matcher"))

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
				return
			}

			if !allowed {
				c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too Many Requests"})
				logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Method(), c.Request.RequestURI(), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
				return
			}
		}

		var (
			rawPath string
			errMsg  string
		)

		rawPath = strings.TrimPrefix(string(c.Request.RequestURI()), "/") // 去掉前缀/

		var (
			user    string
			repo    string
			matcher string
			err     error
		)

		user = c.Param("user")
		repo = c.Param("repo")
		matcher = ctx.Value("matcher").(string)

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
		// dump log 记录详细信息 c.ClientIP(), c.Method(), rawPath,c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), full Header
		logDump("%s %s %s %s %s %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), c.Request.Header.Header())

		// 白名单检查
		if cfg.Whitelist.Enabled {
			var whitelist bool
			whitelist = auth.CheckWhitelist(user, repo)
			if !whitelist {
				errMsg = fmt.Sprintf("Whitelist Blocked repo: %s/%s", user, repo)
				c.JSON(http.StatusForbidden, map[string]string{"error": errMsg})
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
				c.JSON(http.StatusForbidden, map[string]string{"error": errMsg})
				logWarning("%s %s %s %s %s Blacklist Blocked repo: %s/%s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
				return
			}
		}

		if matcher == "api" && !cfg.Auth.ForceAllowApi {
			if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
				c.JSON(http.StatusForbidden, map[string]string{"error": "Github API Req without AuthHeader is Not Allowed"})
				logWarning("%s %s %s %s %s AuthHeader Unavailable", c.ClientIP(), c.Method(), rawPath)
				return
			}
		}

		// 鉴权
		if cfg.Auth.Enabled {
			var authcheck bool
			authcheck, err = auth.AuthHandler(ctx, c, cfg)
			if !authcheck {
				//c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
				c.AbortWithStatusJSON(401, map[string]string{"error": "Unauthorized"})
				logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), err)
				return
			}
		}

		// 若匹配api.github.com/repos/用户名/仓库名/路径, 则检查是否开启HeaderAuth

		// 处理blob/raw路径
		if matcher == "blob" {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		// 为rawpath加入https:// 头
		rawPath = "https://" + rawPath

		// IP METHOD URL USERAGENT PROTO MATCHES
		logDebug("%s %s %s %s %s Matched: %v", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), matcher)

		switch matcher {
		case "releases", "blob", "raw", "gist", "api":
			ChunkedProxyRequest(ctx, c, rawPath, cfg, matcher)
		case "clone":
			GitReq(ctx, c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			fmt.Println("Invalid input.")
			return
		}
	}
}
