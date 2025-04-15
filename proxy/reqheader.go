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

// removeWSHeader removes the "Upgrade" and "Connection" headers from the given
// Request, which are added by the client when it wants to upgrade the
// connection to a WebSocket connection.
func removeWSHeader(req *http.Request) {
	req.Header.Del("Upgrade")
	req.Header.Del("Connection")
}
