// proxy/proxy.go 实验性
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
	regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+`),
}

func NoRouteHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
		matches := re.FindStringSubmatch(rawPath)

		if len(matches) < 3 {
			logWarning("Invalid URL: %s", rawPath)
			c.String(http.StatusForbidden, "Invalid URL.")
			return
		}

		rawPath = "https://" + matches[2]

		username, repo := MatchUserRepo(rawPath, cfg, c, matches)

		logWarning("Blacklist Check > Username: %s, Repo: %s", username, repo)
		fullrepo := fmt.Sprintf("%s/%s", username, repo)

		// 白名单检查
		if cfg.Whitelist.Enabled {
			whitelistpass := auth.CheckWhitelist(fullrepo)
			if !whitelistpass {
				errMsg := fmt.Sprintf("Whitelist Blocked repo: %s", fullrepo)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(errMsg)
				return
			}
		}

		// 黑名单检查
		if cfg.Blacklist.Enabled {
			blacklistpass := auth.CheckBlacklist(fullrepo)
			if blacklistpass {
				errMsg := fmt.Sprintf("Blacklist Blocked repo: %s", fullrepo)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(errMsg)
				return
			}
		}

		matches = CheckURL(rawPath)
		if matches == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if exps[1].MatchString(rawPath) {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		if !auth.AuthHandler(c, cfg) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			logWarning("Unauthorized request: %s", rawPath)
			return
		}

		logInfo("Matches: %v", matches)

		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			logInfo("%s Matched - USE proxy-chrome", rawPath)
			ProxyRequest(c, rawPath, cfg, "chrome")
		case exps[2].MatchString(rawPath):
			logInfo("%s Matched - USE proxy-git", rawPath)
			ProxyRequest(c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}
	}
}

func MatchUserRepo(rawPath string, cfg *config.Config, c *gin.Context, matches []string) (string, string) {
	// 提取用户名和仓库名，格式为 handle/<username>/<repo>/*
	pathmatches := regexp.MustCompile(`^([^/]+)/([^/]+)/([^/]+)/.*`)
	pathParts := pathmatches.FindStringSubmatch(matches[2])

	if len(pathParts) < 4 {
		logWarning("Invalid path: %s", rawPath)
		c.String(http.StatusForbidden, "Invalid path; expected username/repo.")
		return "", ""
	} else {
		return pathParts[2], pathParts[3]
	}
}

func ProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string) {
	method := c.Request.Method
	// 记录日志 IP 地址、请求方法、请求 URL、请求头 User-Agent 、HTTP版本
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	client := createHTTPClient(mode)

	body, err := readRequestBody(c)
	if err != nil {
		HandleError(c, err.Error())
		return
	}

	req := client.R().SetBody(body)
	setRequestHeaders(c, req)

	resp, err := SendRequest(req, method, u)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := HandleResponseSize(resp, cfg, c); err != nil {
		logWarning("Error handling response size: %v", err)
		return
	}

	CopyResponseHeaders(resp, c, cfg)
	c.Status(resp.StatusCode)
	if err := copyResponseBody(c, resp.Body); err != nil {
		logError("Failed to copy response body: %v", err)
	}
}

func createHTTPClient(mode string) *req.Client {
	client := req.C()
	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome()
	case "git":
		client.SetUserAgent("git/2.33.1")
	}
	return client
}

// readRequestBody 读取请求体
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer c.Request.Body.Close()
	return body, nil
}

// setRequestHeaders 设置请求头
func setRequestHeaders(c *gin.Context, req *req.Request) {
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}
}

// copyResponseBody 复制响应体到客户端
func copyResponseBody(c *gin.Context, respBody io.Reader) error {
	_, err := io.Copy(c.Writer, respBody)
	return err
}

func SendRequest(req *req.Request, method, url string) (*req.Response, error) {
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
		logInfo("Unsupported method: %s", method)
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func HandleResponseSize(resp *req.Response, cfg *config.Config, c *gin.Context) error {
	contentLength := resp.Header.Get("Content-Length")
	sizelimit := cfg.Server.SizeLimit * 1024 * 1024
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > sizelimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logWarning("Size limit exceeded: %s, Size: %d", finalURL, size)
			return fmt.Errorf("size limit exceeded: %d", size)
		}
	}
	return nil
}

func CopyResponseHeaders(resp *req.Response, c *gin.Context, cfg *config.Config) {

	copyHeaders(resp, c)

	removeHeaders(resp)

	setCORSHeaders(c, cfg)

	setDefaultHeaders(c)
}

// removeHeaders 移除指定的响应头
func removeHeaders(resp *req.Response) {
	headersToRemove := map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
	}

	for header := range headersToRemove {
		resp.Header.Del(header)
	}
}

// copyHeaders 复制响应头到 Gin 上下文
func copyHeaders(resp *req.Response, c *gin.Context) {
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
}

// setCORSHeaders 设置 CORS 相关的响应头
func setCORSHeaders(c *gin.Context, cfg *config.Config) {
	if cfg.CORS.Enabled {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}
}

// setDefaultHeaders 设置默认的响应头
func setDefaultHeaders(c *gin.Context) {
	c.Header("Age", "10")
	c.Header("Cache-Control", "max-age=300")
}

func HandleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", message))
	logWarning(message)
}

func CheckURL(u string) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			logInfo("URL matched: %s, Matches: %v", u, matches[1:])
			return matches[1:]
		}
	}
	errMsg := fmt.Sprintf("Invalid URL: %s", u)
	logWarning(errMsg)
	return nil
}
