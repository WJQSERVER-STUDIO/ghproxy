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

var logw = logger.Logw

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
			logw("Invalid URL: %s", rawPath)
			c.String(http.StatusForbidden, "Invalid URL.")
			return
		}

		rawPath = "https://" + matches[2]

		// 提取用户名和仓库名，格式为 handle/<username>/<repo>/*
		pathmatches := regexp.MustCompile(`^([^/]+)/([^/]+)/([^/]+)/.*`)
		pathParts := pathmatches.FindStringSubmatch(matches[2])
		if len(pathParts) < 4 {
			logw("Invalid path: %s", rawPath)
			c.String(http.StatusForbidden, "Invalid path; expected username/repo.")
			return
		}

		username := pathParts[2]
		repo := pathParts[3]
		logw("Blacklist Check > Username: %s, Repo: %s", username, repo)
		fullrepo := fmt.Sprintf("%s/%s", username, repo)

		// 白名单检查
		if cfg.Whitelist.Enabled {
			whitelistpass := auth.CheckWhitelist(fullrepo)
			if !whitelistpass {
				errMsg := fmt.Sprintf("Whitelist Blocked repo: %s", fullrepo)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logw(errMsg)
				return
			}
		}

		// 黑名单检查
		if cfg.Blacklist.Enabled {
			blacklistpass := auth.CheckBlacklist(fullrepo)
			if blacklistpass {
				errMsg := fmt.Sprintf("Blacklist Blocked repo: %s", fullrepo)
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				logw(errMsg)
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
			logw("Unauthorized request: %s", rawPath)
			return
		}

		logw("Matches: %v", matches)

		switch {
		case exps[0].MatchString(rawPath), exps[1].MatchString(rawPath), exps[3].MatchString(rawPath), exps[4].MatchString(rawPath):
			logw("%s Matched - USE proxy-chrome", rawPath)
			ProxyRequest(c, rawPath, cfg, "chrome")
		case exps[2].MatchString(rawPath):
			logw("%s Matched - USE proxy-git", rawPath)
			ProxyRequest(c, rawPath, cfg, "git")
		default:
			c.String(http.StatusForbidden, "Invalid input.")
			return
		}
	}
}

func ProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string) {
	method := c.Request.Method
	logw("%s %s", method, u)

	client := req.C()

	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome()
	case "git":
		client.SetUserAgent("git/2.33.1")
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer c.Request.Body.Close()

	req := client.R().SetBody(body)

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}

	resp, err := SendRequest(req, method, u)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := HandleResponseSize(resp, cfg, c); err != nil {
		logw("Error handling response size: %v", err)
		return
	}

	CopyResponseHeaders(resp, c, cfg)
	c.Status(resp.StatusCode)
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		logw("Failed to copy response body: %v", err)
	}
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
		logw("Unsupported method: %s", method)
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func HandleResponseSize(resp *req.Response, cfg *config.Config, c *gin.Context) error {
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > cfg.Server.SizeLimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logw("Redirecting to %s due to size limit (%d bytes)", finalURL, size)
			return fmt.Errorf("response size exceeds limit")
		}
	}
	return nil
}

func CopyResponseHeaders(resp *req.Response, c *gin.Context, cfg *config.Config) {
	headersToRemove := map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
	}

	for header := range headersToRemove {
		resp.Header.Del(header)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Header("Access-Control-Allow-Origin", "")
	if cfg.CORS.Enabled {
		c.Header("Access-Control-Allow-Origin", "*")
	}

	c.Header("Age", "10")
	c.Header("Cache-Control", "max-age=300")
}

func HandleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", message))
	logw(message)
}

func CheckURL(u string) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			logw("URL matched: %s, Matches: %v", u, matches[1:])
			return matches[1:]
		}
	}
	errMsg := fmt.Sprintf("Invalid URL: %s", u)
	logw(errMsg)
	return nil
}
