package proxy

import (
	"bytes"
	"context"
	"fmt"
	"ghproxy/config"
	"net/http"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/cloudwego/hertz/pkg/app"
)

func GitReq(ctx context.Context, c *app.RequestContext, u string, cfg *config.Config, mode string) {

	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	go func() {
		<-ctx.Done()
		if resp != nil && resp.Body != nil {
			err = resp.Body.Close()
			if err != nil {
				logError("Failed to close response body: %v", err)
			}
		}
	}()

	method := string(c.Request.Method())

	reqBodyReader := bytes.NewBuffer(c.Request.Body())

	//bodyReader := c.Request.BodyStream() // 不可替换为此实现

	if cfg.GitClone.Mode == "cache" {
		userPath, repoPath, remainingPath, queryParams, err := extractParts(u)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to extract parts from URL: %v", err))
			return
		}
		// 构建新url
		u = cfg.GitClone.SmartGitAddr + userPath + repoPath + remainingPath + "?" + queryParams.Encode()
	}

	if cfg.GitClone.Mode == "cache" {
		rb := gitclient.NewRequestBuilder(method, u)
		rb.NoDefaultHeaders()
		rb.SetBody(reqBodyReader)
		rb.WithContext(ctx)

		req, err = rb.Build()
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
			return
		}

		setRequestHeaders(c, req, cfg, "clone")
		AuthPassThrough(c, cfg, req)

		resp, err = gitclient.Do(req)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
			return
		}
	} else {
		rb := client.NewRequestBuilder(string(c.Request.Method()), u)
		rb.NoDefaultHeaders()
		rb.SetBody(reqBodyReader)
		rb.WithContext(ctx)

		req, err := rb.Build()
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
			return
		}

		setRequestHeaders(c, req, cfg, "clone")
		AuthPassThrough(c, cfg, req)

		resp, err = client.Do(req)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
			return
		}
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		sizelimit := cfg.Server.SizeLimit * 1024 * 1024
		if err != nil {
			logWarning("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), err)
		}
		if err == nil && size > sizelimit {
			finalURL := []byte(resp.Request.URL.String())
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol(), finalURL, size)
			return
		}
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Response.Header.Add(key, value)
		}
	}

	headersToRemove := map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
	}

	for header := range headersToRemove {
		resp.Header.Del(header)
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
	if cfg.GitClone.Mode == "cache" {
		c.Response.Header.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Response.Header.Set("Pragma", "no-cache")
		c.Response.Header.Set("Expires", "0")
	}

	bodyReader := resp.Body

	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	c.SetBodyStream(bodyReader, -1)
}
