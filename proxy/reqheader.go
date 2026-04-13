package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/infinite-iroha/touka"
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
		"Accept-Encoding":  {},
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
		"Accept-Encoding":   "",
		"Transfer-Encoding": "chunked",
		"User-Agent":        "GHProxy/1.0",
	}
)

// copyHeader 将所有头部从 src 复制到 dst。
// 对于多值头部，它会为每个值调用 Add，从而保留所有值。
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func canonicalizeHeaderSet(headers map[string]struct{}) map[string]struct{} {
	canonicalized := make(map[string]struct{}, len(headers))
	for key := range headers {
		canonicalized[http.CanonicalHeaderKey(key)] = struct{}{}
	}
	return canonicalized
}

func init() {
	reqHeadersToRemove = canonicalizeHeaderSet(reqHeadersToRemove)
	cloneHeadersToRemove = canonicalizeHeaderSet(cloneHeadersToRemove)
	respHeadersToRemove = canonicalizeHeaderSet(respHeadersToRemove)
	defaultHeaders = map[string]string{
		"Accept":            "*/*",
		"Accept-Encoding":   "",
		"Transfer-Encoding": "chunked",
		"User-Agent":        "GHProxy/1.0",
	}
}

func copyHeaderFiltered(dst, src http.Header, denylist map[string]struct{}) {
	for k, vv := range src {
		if _, denied := denylist[k]; denied {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func setRequestHeaders(c *touka.Context, req *http.Request, cfg *config.Config, matcher string) {
	if matcher == "raw" && cfg.Httpc.UseCustomRawHeaders {
		// 使用预定义Header
		for key, value := range defaultHeaders {
			req.Header.Set(key, value)
		}
	} else if matcher == "clone" {
		copyHeaderFiltered(req.Header, c.Request.Header, cloneHeadersToRemove)
	} else {
		copyHeaderFiltered(req.Header, c.Request.Header, reqHeadersToRemove)
	}
}
