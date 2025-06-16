package proxy

import (
	"context"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/infinite-iroha/touka"
)

func ChunkedProxyRequest(ctx context.Context, c *touka.Context, u string, cfg *config.Config, matcher string) {

	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	go func() {
		<-ctx.Done()
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if req != nil && req.Body != nil {
			req.Body.Close()
		}
	}()

	rb := client.NewRequestBuilder(c.Request.Method, u)
	rb.NoDefaultHeaders()
	rb.SetBody(c.Request.Body)
	rb.WithContext(ctx)

	req, err = rb.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	setRequestHeaders(c, req, cfg, matcher)
	AuthPassThrough(c, cfg, req)

	resp, err = client.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}

	// 错误处理(404)
	if resp.StatusCode == 404 {
		ErrorPage(c, NewErrorWithStatusLookup(404, "Page Not Found (From Github)"))
		return
	}

	// 处理302情况
	if resp.StatusCode == 302 || resp.StatusCode == 301 {
		//c.Debugf("resp header %s", resp.Header)
		finalURL := resp.Header.Get("Location")
		if finalURL != "" {
			err = resp.Body.Close()
			if err != nil {
				c.Errorf("Failed to close response body: %v", err)
			}
			c.Infof("Internal Redirecting to %s", finalURL)
			ChunkedProxyRequest(ctx, c, finalURL, cfg, matcher)
			return
		}
	}

	// 处理响应体大小限制

	var (
		bodySize      int
		contentLength string
		sizelimit     int
	)
	sizelimit = cfg.Server.SizeLimit * 1024 * 1024
	contentLength = resp.Header.Get("Content-Length")
	if contentLength != "" {
		var err error
		bodySize, err = strconv.Atoi(contentLength)
		if err != nil {
			c.Warnf("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, err)
			bodySize = -1
		}
		if err == nil && bodySize > sizelimit {
			finalURL := resp.Request.URL.String()
			err = resp.Body.Close()
			if err != nil {
				c.Errorf("Failed to close response body: %v", err)
			}
			c.Redirect(301, finalURL)
			c.Warnf("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, finalURL, bodySize)
			return
		}
	}

	// 复制响应头，排除需要移除的 header
	for key, values := range resp.Header {
		if _, shouldRemove := respHeadersToRemove[key]; !shouldRemove {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}

	switch cfg.Server.Cors {
	case "*":
		c.Header("Access-Control-Allow-Origin", "*")
	case "":
		c.Header("Access-Control-Allow-Origin", "*")
	case "nil":
		c.Header("Access-Control-Allow-Origin", "")
	default:
		c.Header("Access-Control-Allow-Origin", cfg.Server.Cors)
	}

	c.Status(resp.StatusCode)

	bodyReader := resp.Body

	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	defer bodyReader.Close()

	if MatcherShell(u) && matchString(matcher) && cfg.Shell.Editor {
		// 判断body是不是gzip
		var compress string
		if resp.Header.Get("Content-Encoding") == "gzip" {
			compress = "gzip"
		}

		c.Debugf("Use Shell Editor: %s %s %s %s %s", c.ClientIP(), c.Request.Method, u, c.UserAgent(), c.Request.Proto)
		c.Header("Content-Length", "")

		var reader io.Reader

		reader, _, err = processLinks(bodyReader, compress, c.Request.Host, cfg, c)
		c.WriteStream(reader)
		if err != nil {
			c.Errorf("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), c.Request.Method, u, c.UserAgent(), c.Request.Proto, err)
			ErrorPage(c, NewErrorWithStatusLookup(500, fmt.Sprintf("Failed to copy response body: %v", err)))
			return
		}
	} else {

		if contentLength != "" {
			c.SetHeader("Content-Length", contentLength)
			c.WriteStream(bodyReader)
			return
		}
		c.WriteStream(bodyReader)
	}

}
