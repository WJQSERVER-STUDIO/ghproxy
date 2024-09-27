package proxy

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
)

var (
	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+`),
	}
)

// var cfg *config.Config
var logw = logger.Logw

func NoRouteHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
		matches := re.FindStringSubmatch(rawPath)

		if len(matches) < 3 {
			c.String(http.StatusBadRequest, "Invalid URL format.")
			return
		}

		rawPath = "https://" + matches[2]
		matches = checkURL(rawPath)
		if matches == nil {
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}

		if exps[1].MatchString(rawPath) {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		if !auth.AuthHandler(c) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			logw("Unauthorized request: %s", rawPath)
			return
		}

		logw("Request: %s %s", c.Request.Method, rawPath)
		logw("Matches: %v", matches)

		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			logw("%s Matched - USE proxy-chrome", rawPath)
			proxyRequest(c, rawPath, cfg, "chrome")
		case exps[2].MatchString(rawPath):
			logw("%s Matched - USE proxy-git", rawPath)
			proxyRequest(c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}
	}
}

func proxyRequest(c *gin.Context, u string, cfg *config.Config, mode string) {
	method := c.Request.Method
	logw("%s Method: %s", u, method)

	client := req.C()

	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36").SetTLSFingerprintChrome().ImpersonateChrome()
	case "git":
		client.SetUserAgent("git/2.33.1")
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		handleError(c, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer c.Request.Body.Close()

	if err := c.Request.Body.Close(); err != nil {
		logw("Failed to close request body: %v", err)
		return
	}

	// 创建新的请求
	req := client.R().SetBody(body)

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}

	// 发送请求并处理响应
	resp, err := sendRequest(req, method, u)
	if err != nil {
		handleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := handleResponseSize(resp, cfg, c); err != nil {
		logw("Error handling response size: %v", err)
		return
	}

	copyResponseHeaders(resp, c, cfg)
	c.Status(resp.StatusCode)
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		logw("Failed to copy response body: %v", err)
	}
}

func sendRequest(req *req.Request, method, url string) (*req.Response, error) {
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
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func handleResponseSize(resp *req.Response, cfg *config.Config, c *gin.Context) error {
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > cfg.SizeLimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusTemporaryRedirect, finalURL) // 改为临时重定向
			logw("Redirecting to %s due to size limit (%d bytes)", finalURL, size)
			return fmt.Errorf("response size exceeds limit")
		}
	}
	return nil
}

func copyResponseHeaders(resp *req.Response, c *gin.Context, cfg *config.Config) {
	headersToRemove := []string{"Content-Security-Policy", "Referrer-Policy", "Strict-Transport-Security"}

	for _, header := range headersToRemove {
		resp.Header.Del(header)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	if cfg.CORSOrigin {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}
}

func handleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("Server error: %v", message))
	logw(message)
}

func checkURL(u string) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			logw("URL matched: %s, Matches: %v", u, matches[1:])
			return matches[1:]
		}
	}
	logw("Invalid URL: %s", u)
	return nil
}

/*func AuthHandler(c *gin.Context) bool {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth {
		logw("auth PASS")
		return true
	}

	// 获取 auth_token 参数
	authToken := c.Query("auth_token")
	logw("auth_token: ", authToken)

	// 验证 token
	isValid := authToken == cfg.AuthToken
	if !isValid {
		logw("auth FAIL")
	}

	return isValid
}*/
