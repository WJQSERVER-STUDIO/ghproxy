package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/rate"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func NoRouteHandler(cfg *config.Config, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter, runMode string) gin.HandlerFunc {
	return func(c *gin.Context) {

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
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too Many Requests"})
				logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Request.Method, c.Request.URL.RequestURI(), c.Request.Header.Get("User-Agent"), c.Request.Proto)
				return
			}
		}

		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/") // 去掉前缀/
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)           // 匹配http://或https://开头的路径
		matches := re.FindStringSubmatch(rawPath)                      // 匹配路径

		// 匹配路径错误处理
		if len(matches) < 3 {
			errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
			logWarning(errMsg)
			c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)
			return
		}

		// 制作url
		rawPath = "https://" + matches[2]

		username, repo := MatchUserRepo(rawPath, cfg, c, matches) // 匹配用户名和仓库名

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, username, repo)
		// dump log 记录详细信息 c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, full Header
		LogDump("%s %s %s %s %s %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, c.Request.Header)
		repouser := fmt.Sprintf("%s/%s", username, repo)

		// 白名单检查
		if cfg.Whitelist.Enabled {
			whitelist := auth.CheckWhitelist(repouser, username, repo)
			if !whitelist {
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Whitelist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser)
				errMsg := fmt.Sprintf("Whitelist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(logErrMsg)
				return
			}
		}

		// 黑名单检查
		if cfg.Blacklist.Enabled {
			blacklist := auth.CheckBlacklist(username, repo)
			if blacklist {
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Blacklist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser)
				errMsg := fmt.Sprintf("Blacklist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(logErrMsg)
				return
			}
		}

		matches = CheckURL(rawPath, c)
		if matches == nil {
			c.AbortWithStatus(http.StatusNotFound)
			logWarning("%s %s %s %s %s 404-NOMATCH", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
			return
		}

		// 若匹配api.github.com/repos/用户名/仓库名/路径, 则检查是否开启HeaderAuth
		if exps[5].MatchString(rawPath) {
			if cfg.Auth.AuthMethod != "header" || !cfg.Auth.Enabled {
				c.JSON(http.StatusForbidden, gin.H{"error": "HeaderAuth is not enabled."})
				logError("%s %s %s %s %s HeaderAuth-Error: HeaderAuth is not enabled.", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
				return
			}
		}

		// 处理blob/raw路径
		if exps[1].MatchString(rawPath) {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		// 鉴权
		authcheck, err := auth.AuthHandler(c, cfg)
		if !authcheck {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		}

		// IP METHOD URL USERAGENT PROTO MATCHES
		logDebug("%s %s %s %s %s Matches: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, matches)

		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			//ProxyRequest(c, rawPath, cfg, "chrome", runMode)
			ChunkedProxyRequest(c, rawPath, cfg, "chrome", runMode) // dev test chunk
		case exps[2].MatchString(rawPath):
			//ProxyRequest(c, rawPath, cfg, "git", runMode)
			GitReq(c, rawPath, cfg, "git", runMode)
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			fmt.Println("Invalid input.")
			return
		}
	}
}
