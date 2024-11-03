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
	"ghproxy/rate"

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
	regexp.MustCompile(`^(?:https?://)?gist\.github(?:usercontent|)\.com/([^/]+)/.+?/.+`),
}

func NoRouteHandler(cfg *config.Config, limiter *rate.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 限制访问频率
		if cfg.RateLimit.Enabled {
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too Many Requests"})
				logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Request.Method, c.Request.URL.RequestURI(), c.Request.Header.Get("User-Agent"), c.Request.Proto)
				return
			}
		}

		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
		matches := re.FindStringSubmatch(rawPath)

		if len(matches) < 3 {
			errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
			logWarning(errMsg)
			c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)
			return
		}

		rawPath = "https://" + matches[2]

		username, repo := MatchUserRepo(rawPath, cfg, c, matches)

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, username, repo)
		repouser := fmt.Sprintf("%s/%s", username, repo)

		// 白名单检查
		if cfg.Whitelist.Enabled {
			whitelist := auth.CheckWhitelist(repouser)
			if !whitelist {
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Whitelist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser)
				errMsg := fmt.Sprintf("Whitelist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(logErrMsg)
				return
			}
		}

		// 黑名单检查
		if cfg.Blacklist.Enabled {
			blacklist := auth.CheckBlacklist(repouser, username, repo)
			if blacklist {
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Whitelist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser)
				errMsg := fmt.Sprintf("Blacklist Blocked repo: %s", repouser)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logWarning(logErrMsg)
				return
			}
		}

		matches = CheckURL(rawPath, c)
		if matches == nil {
			c.AbortWithStatus(http.StatusNotFound)
			logError("%s %s %s %s %s 404-NOMATCH", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto)
			return
		}

		if exps[1].MatchString(rawPath) {
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
		}

		// 鉴权
		authcheck, err := auth.AuthHandler(c, cfg)
		if !authcheck {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		}

		// IP METHOD URL USERAGENT PROTO MATCHES
		logInfo("%s %s %s %s %s Matches: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, matches)

		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			ProxyRequest(c, rawPath, cfg, "chrome")
		case exps[2].MatchString(rawPath):
			ProxyRequest(c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			fmt.Println("Invalid input.")
			return
		}
	}
}

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

func ProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	client := createHTTPClient(mode)

	body, err := readRequestBody(c)
	if err != nil {
		HandleError(c, err.Error())
		return
	}

	req := client.R().SetBody(body)
	setRequestHeaders(c, req)

	resp, err := SendRequest(c, req, method, u)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := HandleResponseSize(resp, cfg, c); err != nil {
		logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	}

	CopyResponseHeaders(resp, c, cfg)
	c.Status(resp.StatusCode)
	if err := copyResponseBody(c, resp.Body); err != nil {
		logError("%s %s %s %s %s Response-Copy-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
	}
}

// 判断并选择TLS指纹
func createHTTPClient(mode string) *req.Client {
	client := req.C()
	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome()
	case "git":
		client.SetUserAgent("git/2.33.1")
	}
	return client
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

// 设置请求头
func setRequestHeaders(c *gin.Context, req *req.Request) {
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}
}

// 复制响应体
func copyResponseBody(c *gin.Context, respBody io.Reader) error {
	_, err := io.Copy(c.Writer, respBody)
	return err
}

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

func HandleResponseSize(resp *req.Response, cfg *config.Config, c *gin.Context) error {
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

func CopyResponseHeaders(resp *req.Response, c *gin.Context, cfg *config.Config) {

	copyHeaders(resp, c)

	removeHeaders(resp)

	setCORSHeaders(c, cfg)

	setDefaultHeaders(c)
}

// 移除指定响应头
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

// 复制响应头
func copyHeaders(resp *req.Response, c *gin.Context) {
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
}

// CORS配置
func setCORSHeaders(c *gin.Context, cfg *config.Config) {
	if cfg.CORS.Enabled {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}
}

// 默认响应
func setDefaultHeaders(c *gin.Context) {
	c.Header("Age", "10")
	c.Header("Cache-Control", "max-age=300")
}

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
