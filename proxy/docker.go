package proxy

import (
	"context"
	"fmt"
	"ghproxy/config"
	"net/http"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/cloudwego/hertz/pkg/app"
)

func GhcrRouting(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if cfg.Docker.Enabled {
			if cfg.Docker.Target == "ghcr" {
				GhcrRequest(ctx, c, "https://ghcr.io"+string(c.Request.RequestURI()), cfg, "ghcr")
			} else if cfg.Docker.Target == "dockerhub" {
				GhcrRequest(ctx, c, "https://registry-1.docker.io"+string(c.Request.RequestURI()), cfg, "dockerhub")
			} else if cfg.Docker.Target != "" {
				// 自定义taget
				GhcrRequest(ctx, c, "https://"+cfg.Docker.Target+string(c.Request.RequestURI()), cfg, "custom")
			} else {
				// 配置为空
				ErrorPage(c, NewErrorWithStatusLookup(403, "Docker Target is not set"))
				return
			}

		} else {
			ErrorPage(c, NewErrorWithStatusLookup(403, "Docker is not Allowed"))
			return
		}
	}
}

func GhcrRequest(ctx context.Context, c *app.RequestContext, u string, cfg *config.Config, matcher string) {

	var (
		method []byte
		req    *http.Request
		resp   *http.Response
		err    error
	)

	go func() {
		<-ctx.Done()
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if req != nil {
			req.Body.Close()
		}
	}()

	method = c.Request.Method()

	rb := client.NewRequestBuilder(string(method), u)
	rb.NoDefaultHeaders()
	rb.SetBody(c.Request.BodyStream())
	rb.WithContext(ctx)

	//req, err = client.NewRequest(string(method), u, c.Request.BodyStream())
	req, err = rb.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	c.Request.Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		headerValue := string(value)
		req.Header.Add(headerKey, headerValue)
	})

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
			logWarning("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), err)
			bodySize = -1
		}
		if err == nil && bodySize > sizelimit {
			var finalURL string
			finalURL = resp.Request.URL.String()
			err = resp.Body.Close()
			if err != nil {
				logError("Failed to close response body: %v", err)
			}
			c.Redirect(301, []byte(finalURL))
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), finalURL, bodySize)
			return
		}
	}

	// 复制响应头，排除需要移除的 header
	for key, values := range resp.Header {
		for _, value := range values {
			//c.Header(key, value)
			c.Response.Header.Add(key, value)
		}
	}

	c.Status(resp.StatusCode)

	bodyReader := resp.Body

	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	if contentLength != "" {
		c.SetBodyStream(bodyReader, bodySize)
		return
	}
	c.SetBodyStream(bodyReader, -1)

}
