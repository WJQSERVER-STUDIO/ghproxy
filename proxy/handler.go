package proxy

import (
	"fmt"
	"ghproxy/config"
	"regexp"
	"strings"

	"github.com/infinite-iroha/touka"
)

// buildHandlerPath 使用 strings.Builder 来高效地构建最终的 URL.
// 这避免了使用标准字符串拼接时发生的多次内存分配.
func buildHandlerPath(path, matcher string) string {
	var sb strings.Builder
	sb.Grow(len(path) + 50)

	if matcher == "blob" && strings.HasPrefix(path, "github.com") {
		sb.WriteString("https://raw.githubusercontent.com")
		if len(path) > 10 { // len("github.com")
			pathSegment := path[10:] // skip "github.com"
			if i := strings.Index(pathSegment, "/blob/"); i != -1 {
				sb.WriteString(pathSegment[:i])
				sb.WriteString("/")
				sb.WriteString(pathSegment[i+len("/blob/"):])
			} else {
				sb.WriteString(pathSegment)
			}
		}
	} else {
		sb.WriteString("https://")
		sb.WriteString(path)
	}

	return sb.String()
}

var re = regexp.MustCompile(`^(http:|https:)?/?/?(.*)`) // 匹配http://或https://开头的路径

func NoRouteHandler(cfg *config.Config) touka.HandlerFunc {
	return func(c *touka.Context) {
		var ctx = c.Request.Context()
		var shoudBreak bool

		var (
			rawPath string
			matches []string
		)

		rawPath = strings.TrimPrefix(c.GetRequestURI(), "/") // 去掉前缀/
		matches = re.FindStringSubmatch(rawPath)             // 匹配路径

		// 匹配路径错误处理
		if len(matches) < 3 {
			c.Warnf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto)
			ErrorPage(c, NewErrorWithStatusLookup(400, fmt.Sprintf("Invalid URL Format: %s", c.GetRequestURI())))
			return
		}

		// 制作url
		path := matches[2]
		var matcherErr *GHProxyErrors
		user, repo, matcher, matcherErr := Matcher("https://"+path, cfg)
		if matcherErr != nil {
			ErrorPage(c, matcherErr)
			return
		}

		rawPath = buildHandlerPath(path, matcher)

		shoudBreak = listCheck(cfg, c, user, repo, rawPath)
		if shoudBreak {
			return
		}

		shoudBreak = authCheck(c, cfg, matcher, rawPath)
		if shoudBreak {
			return
		}

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
