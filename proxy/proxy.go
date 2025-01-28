package proxy

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-gonic/gin"
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

var exps = []*regexp.Regexp{
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*`),
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*`),
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*`),
	regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+`),
	regexp.MustCompile(`^(?:https?://)?gist\.github(?:usercontent|)\.com/([^/]+)/.+?/.+`),
	regexp.MustCompile(`^(?:https?://)?api\.github\.com/repos/([^/]+)/([^/]+)/.*`),
}

// 读取请求体
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer c.Request.Body.Close()
	return body, nil
}

/*
func SendRequest(c *gin.Context, req *req.Request, method, url string) (*req.Response, error) {
	switch method {
	case "GET":
		return req.Get(url)
	case "POST":
		return req.Post(url)
	case "PUT":
		return req.Put(url)
	case "DELETE":
		return req.Delete(url)
	default:
		// IP METHOD URL USERAGENT PROTO UNSUPPORTED-METHOD
		errmsg := fmt.Sprintf("%s %s %s %s %s Unsupported method", c.ClientIP(), method, url, c.Request.Header.Get("User-Agent"), c.Request.Proto)
		logWarning(errmsg)
		return nil, fmt.Errorf(errmsg)
	}
}
*/

func HandleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", message))
	logWarning(message)
}

func CheckURL(u string, c *gin.Context) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			return matches[1:]
		}
	}
	errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)
	logWarning(errMsg)
	return nil
}

/*
// 处理响应大小
func HandleResponseSize(resp *http.Response, cfg *config.Config, c *gin.Context) error {
	contentLength := resp.Header.Get("Content-Length")
	sizelimit := cfg.Server.SizeLimit * 1024 * 1024
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > sizelimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto, finalURL, size)
			return fmt.Errorf("Path: %s size limit exceeded: %d", finalURL, size)
		}
	}
	return nil
}
*/
