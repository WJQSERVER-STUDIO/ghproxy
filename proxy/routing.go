package proxy

import (
	"context"
	"ghproxy/config"
	"ghproxy/rate"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

func RoutingHandler(cfg *config.Config, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		var shoudBreak bool

		shoudBreak = rateCheck(cfg, c, limiter, iplimiter)
		if shoudBreak {
			return
		}

		var (
			rawPath string
		)

		rawPath = strings.TrimPrefix(string(c.Request.RequestURI()), "/") // 去掉前缀/

		var (
			user    string
			repo    string
			matcher string
		)

		user = c.Param("user")
		repo = c.Param("repo")
		matcher = c.GetString("matcher")

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
		logDump("%s", c.Request.Header.Header())

		shoudBreak = listCheck(cfg, c, user, repo, rawPath)
		if shoudBreak {
			return
		}

		shoudBreak = authCheck(c, cfg, matcher, rawPath)
		if shoudBreak {
			return
		}

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
			c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid input."})
			logError("Invalid input")
			return
		}
	}
}
