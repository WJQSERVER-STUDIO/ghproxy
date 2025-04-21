package proxy

import (
	"context"
	"errors"
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

		rateCheck(cfg, c, limiter, iplimiter)

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
			//c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)
			c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid URL Format"})
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
				c.JSON(ErrInvalidURL.Code, map[string]string{"error": "Invalid URL Format, Path: " + rawPath})
				logWarning(err.Error())
				return
			}
			if errors.Is(err, ErrAuthHeaderUnavailable) {
				c.JSON(ErrAuthHeaderUnavailable.Code, map[string]string{"error": "AuthHeader Unavailable"})
				logWarning(err.Error())
				return
			}
			if errors.Is(err, ErrNotFound) {
				//c.JSON(ErrNotFound.Code, map[string]string{"error": "Not Found"})
				NotFoundPage(c)
				logWarning(err.Error())
				return
			}
		}

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Method(), rawPath, c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), user, repo)
		logDump("%s", c.Request.Header.Header())

		listCheck(cfg, c, user, repo, rawPath)
		authCheck(c, cfg, matcher, rawPath)

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
