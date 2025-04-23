package proxy

import (
	"context"
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

		var shoudBreak bool
		shoudBreak = rateCheck(cfg, c, limiter, iplimiter)
		if shoudBreak {
			return
		}

		var (
			rawPath string
			matches []string
		)

		rawPath = strings.TrimPrefix(string(c.Request.RequestURI()), "/") // 去掉前缀/
		matches = re.FindStringSubmatch(rawPath)                          // 匹配路径
		logDebug("URL: %v", matches)

		// 匹配路径错误处理
		if len(matches) < 3 {
			logWarning("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol())
			c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid URL Format"})
			return
		}

		// 制作url
		rawPath = "https://" + matches[2]

		var (
			user    string
			repo    string
			matcher string
		)

		var matcherErr *GHProxyErrors
		user, repo, matcher, matcherErr = Matcher(rawPath, cfg)
		if matcherErr != nil {
			ErrorPage(c, matcherErr)
			return
		}

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

		logDebug("Matched: %v", matcher)

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
