package proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 设置请求头
func setRequestHeaders(c *gin.Context, req *http.Request) {
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
}

func removeWSHeader(req *http.Request) {
	req.Header.Del("Upgrade")
	req.Header.Del("Connection")
}
