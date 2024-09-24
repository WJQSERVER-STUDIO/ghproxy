package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
)

var cfg *config.Config
var logw = logger.Logw
var router *gin.Engine

var (
	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+`),
	}
)

func loadConfig() {
	var err error
	// 初始化配置
	cfg, err = config.LoadConfig("/data/go/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger() {
	// 初始化日志模块
	var err error
	err = logger.Init(cfg.LogFilePath) // 传递日志文件路径
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logw("Logger initialized")
	logw("Init Completed")
}

func init() {
	loadConfig()
	setupLogger()

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化路由
	router = gin.Default()

	// 定义路由
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://ghproxy0rtt.1888866.xyz/")
	})

	router.GET("/api", api)

	// 健康检查
	router.GET("/api/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 未匹配路由处理
	router.NoRoute(noRouteHandler(cfg))
}

func main() {
	// 启动服务器
	err := router.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}

	fmt.Println("Program finished")
}

func api(c *gin.Context) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"MaxResponseBodySize": cfg.SizeLimit,
	})
}

func authHandler(c *gin.Context) bool {
	if cfg.Auth {
		authToken := c.Query("auth_token")
		return authToken == cfg.AuthToken
	}
	return true
}

func noRouteHandler(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
		matches := re.FindStringSubmatch(rawPath)

		rawPath = "https://" + matches[2]

		matches = checkURL(rawPath)
		if matches == nil {
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}

		if exps[1].MatchString(rawPath) {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		if !authHandler(c) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			logw("Unauthorized request: %s", rawPath)
			return
		}

		// 日志记录
		logw("Request: %s %s", c.Request.Method, rawPath)
		logw("Matches: %v", matches)

		// 代理请求
		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			logw("%s Matched - USE proxy-chrome", rawPath)
			proxyRequest(c, rawPath, config, "chrome")
		case exps[2].MatchString(rawPath):
			logw("%s Matched - USE proxy-git", rawPath)
			proxyRequest(c, rawPath, config, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}
	}
}

func proxyRequest(c *gin.Context, u string, config *config.Config, mode string) {
	method := c.Request.Method
	logw("%s Method: %s", u, method)

	client := req.C()

	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome()
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

	// 检查响应内容长度并处理重定向
	if err := handleResponseSize(resp, config, c); err != nil {
		logw("Error handling response size: %v", err)
		return
	}

	copyResponseHeaders(resp, c, config)
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

func handleResponseSize(resp *req.Response, config *config.Config, c *gin.Context) error {
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > config.SizeLimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logw("Redirecting to %s due to size limit (%d bytes)", finalURL, size)
			return fmt.Errorf("response size exceeds limit")
		}
	}
	return nil
}

func copyResponseHeaders(resp *req.Response, c *gin.Context, config *config.Config) {
	headersToRemove := []string{"Content-Security-Policy", "Referrer-Policy", "Strict-Transport-Security"}

	for _, header := range headersToRemove {
		resp.Header.Del(header)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	if config.CORSOrigin {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}
}

func handleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", message))
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
