package proxy

import (
	"fmt"
	"ghproxy/config"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// 提取用户名和仓库名
func MatchUserRepo(rawPath string, cfg *config.Config, c *gin.Context, matches []string) (string, string) {
	var gistregex = regexp.MustCompile(`^(?:https?://)?gist\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.*`)
	var gistmatches []string
	if gistregex.MatchString(rawPath) {
		gistmatches = gistregex.FindStringSubmatch(rawPath)
		logInfo("%s %s %s %s %s Matched-Username: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, gistmatches[1])
		return gistmatches[1], ""
	}
	// 定义路径
	pathRegex := regexp.MustCompile(`^([^/]+)/([^/]+)/([^/]+)/.*`)
	if pathMatches := pathRegex.FindStringSubmatch(matches[2]); len(pathMatches) >= 4 {
		return pathMatches[2], pathMatches[3]
	}

	// 返回错误信息
	errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
	logWarning(errMsg)
	c.String(http.StatusForbidden, "Invalid path; expected username/repo, Path: %s", rawPath)
	return "", ""
}
