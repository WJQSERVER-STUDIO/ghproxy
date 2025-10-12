package proxy

import (
	"ghproxy/config"
	"strings"

	"github.com/infinite-iroha/touka"
)

// buildRoutingPath 使用 strings.Builder 来高效地构建最终的 URL.
// 这避免了使用标准字符串拼接时发生的多次内存分配.
func buildRoutingPath(rawPath, matcher string) string {
	var sb strings.Builder
	// 预分配内存以提高性能
	// (This comment is in Chinese as requested by the user)
	sb.Grow(len(rawPath) + 30)
	sb.WriteString("https://")

	if matcher == "blob" {
		sb.WriteString("raw.githubusercontent.com")
		if len(rawPath) > 10 { // len("github.com")
			pathSegment := rawPath[10:]
			if i := strings.Index(pathSegment, "/blob/"); i != -1 {
				sb.WriteString(pathSegment[:i])
				sb.WriteString("/")
				sb.WriteString(pathSegment[i+len("/blob/"):])
			} else {
				sb.WriteString(pathSegment)
			}
		}
	} else {
		sb.WriteString(rawPath)
	}
	return sb.String()
}

func RoutingHandler(cfg *config.Config) touka.HandlerFunc {
	return func(c *touka.Context) {

		var shoudBreak bool

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

		rawPath = buildRoutingPath(rawPath, matcher)
		if matcher == "blob" {
			matcher = "raw"
		}

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
