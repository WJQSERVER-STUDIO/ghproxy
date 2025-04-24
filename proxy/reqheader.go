package proxy

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

func setRequestHeaders(c *app.RequestContext, req *http.Request) {
	c.Request.Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		headerValue := string(value)
		if _, shouldRemove := reqHeadersToRemove[headerKey]; !shouldRemove {
			req.Header.Set(headerKey, headerValue)
		}
	})
}
