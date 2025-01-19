package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CopyResponseHeaders(resp *http.Response, c *gin.Context, cfg *config.Config) {

	copyHeaders(resp, c)

	removeHeaders(resp)

	setCORSHeaders(c, cfg)

	setDefaultHeaders(c)
}

// 复制响应头
func copyHeaders(resp *http.Response, c *gin.Context) {
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
}

// 移除指定响应头
func removeHeaders(resp *http.Response) {
	headersToRemove := map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
	}

	for header := range headersToRemove {
		resp.Header.Del(header)
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
