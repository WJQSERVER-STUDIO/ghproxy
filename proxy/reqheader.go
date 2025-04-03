package proxy

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// 设置请求头
func setRequestHeaders(c *app.RequestContext, req *http.Request) {
	c.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
}

func removeWSHeader(req *http.Request) {
	req.Header.Del("Upgrade")
	req.Header.Del("Connection")
}
