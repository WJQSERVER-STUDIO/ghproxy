package proxy

import (
	"fmt"
	"ghproxy/config"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// 预定义regex
var (
	pathRegex = regexp.MustCompile(`^([^/]+)/([^/]+)/([^/]+)/.*`)                                         // 匹配路径
	gistRegex = regexp.MustCompile(`^(?:https?://)?gist\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.*`) // 匹配gist路径
)

// 提取用户名和仓库名
func MatchUserRepo(rawPath string, cfg *config.Config, c *gin.Context, matches []string) (string, string) {
	if gistMatches := gistRegex.FindStringSubmatch(rawPath); len(gistMatches) == 3 {
		logInfo("%s %s %s %s %s Matched-Username: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, gistMatches[1])
		return gistMatches[1], ""
	}
	// 定义路径
	if pathMatches := pathRegex.FindStringSubmatch(matches[2]); len(pathMatches) >= 4 {
		return pathMatches[2], pathMatches[3]
	}

	// 返回错误信息
	errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
	logWarning(errMsg)
	c.String(http.StatusForbidden, "Invalid path; expected username/repo, Path: %s", rawPath)
	return "", ""
}
