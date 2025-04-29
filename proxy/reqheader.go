package proxy

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// 预定义headers
var (
	defaultHeaders = map[string]string{
		"Accept":            "*/*",
		"Accept-Encoding":   "gzip",
		"Transfer-Encoding": "chunked",
		"User-Agent":        "GHProxy/1.0",
	}
)

func setRequestHeaders(c *app.RequestContext, req *http.Request, matcher string) {
	if matcher == "raw" {
		// 使用预定义Header
		for key, value := range defaultHeaders {
			req.Header.Set(key, value)
		}
	} else {
		c.Request.Header.VisitAll(func(key, value []byte) {
			headerKey := string(key)
			headerValue := string(value)
			if _, shouldRemove := reqHeadersToRemove[headerKey]; !shouldRemove {
				req.Header.Set(headerKey, headerValue)
			}
		})
	}
}
