package proxy

import (
	"net/http"
	"strings"

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

func reWriteEncodeHeader(req *http.Request) {

	if isGzipAccepted(req.Header) {
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
	} else {
		req.Header.Del("Content-Encoding")
		req.Header.Del("Accept-Encoding")
	}

}

// isGzipAccepted 检查 Accept-Encoding 头部中是否包含 gzip
func isGzipAccepted(header http.Header) bool {
	// 获取 Accept-Encoding 的值
	encodings := header["Accept-Encoding"]
	for _, encoding := range encodings {
		// 将 encoding 字符串拆分为多个编码
		for _, enc := range strings.Split(encoding, ",") {
			// 去除空格并检查是否为 gzip
			if strings.TrimSpace(enc) == "gzip" {
				return true
			}
		}
	}
	return false
}
