package proxy

import (
	"ghproxy/config"
	"strings"

	"github.com/infinite-iroha/touka"
)

func RoutingHandler(cfg *config.Config) touka.HandlerFunc {
	return func(c *touka.Context) {

		var shoudBreak bool

		//	shoudBreak = rateCheck(cfg, c, limiter, iplimiter)
		//	if shoudBreak {
		//		return
		//}

		var (
			rawPath string
		)

		rawPath = strings.TrimPrefix(c.GetRequestURI(), "/") // 去掉前缀/

		var (
			user string
			repo string
		)

		user = c.Param("user")
		repo = c.Param("repo")
		matcher, exists := c.GetString("matcher")
		if !exists {
			ErrorPage(c, NewErrorWithStatusLookup(500, "Matcher Not Found in Context"))
			c.Errorf("Matcher Not Found in Context Path: %s", c.GetRequestURIPath())
			return
		}

		ctx := c.Request.Context()

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
			rawPath = rawPath[10:]
			rawPath = "raw.githubusercontent.com" + rawPath
			rawPath = strings.Replace(rawPath, "/blob/", "/", 1)
			matcher = "raw"
		}

		// 为rawpath加入https:// 头
		rawPath = "https://" + rawPath

		switch matcher {
		case "releases", "blob", "raw", "gist", "api":
			ChunkedProxyRequest(ctx, c, rawPath, cfg, matcher)
		case "clone":
			GitReq(ctx, c, rawPath, cfg, "git")
		default:
			ErrorPage(c, NewErrorWithStatusLookup(500, "Matched But Not Matched"))
			c.Errorf("Matched But Not Matched Path: %s rawPath: %s matcher: %s", c.GetRequestURIPath(), rawPath, matcher)
			return
		}
	}
}
