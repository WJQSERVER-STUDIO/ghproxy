package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

var (
	respHeadersToRemove = map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
		"X-Github-Request-Id":       {},
		"X-Timer":                   {},
		"X-Served-By":               {},
		"X-Fastly-Request-Id":       {},
	}

	reqHeadersToRemove = map[string]struct{}{
		"CF-IPCountry":     {},
		"CF-RAY":           {},
		"CF-Visitor":       {},
		"CF-Connecting-IP": {},
		"CF-EW-Via":        {},
		"CDN-Loop":         {},
		"Upgrade":          {},
		"Connection":       {},
	}

	cloneHeadersToRemove = map[string]struct{}{
		"CF-IPCountry":     {},
		"CF-RAY":           {},
		"CF-Visitor":       {},
		"CF-Connecting-IP": {},
		"CF-EW-Via":        {},
		"CDN-Loop":         {},
	}
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

func setRequestHeaders(c *app.RequestContext, req *http.Request, cfg *config.Config, matcher string) {
	if matcher == "raw" && cfg.Httpc.UseCustomRawHeaders {
		// 使用预定义Header
		for key, value := range defaultHeaders {
			req.Header.Set(key, value)
		}
	} else if matcher == "clone" {
		c.Request.Header.VisitAll(func(key, value []byte) {
			headerKey := string(key)
			headerValue := string(value)
			if _, shouldRemove := cloneHeadersToRemove[headerKey]; !shouldRemove {
				req.Header.Set(headerKey, headerValue)
			}
		})
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
