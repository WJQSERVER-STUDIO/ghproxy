package proxy

import (
	"context"
	"fmt"
	"ghproxy/config"
	"net/http"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/infinite-iroha/touka"
)

func GitReq(ctx context.Context, c *touka.Context, u string, cfg *config.Config, mode string) {

	var (
		resp *http.Response
	)

	go func() {
		<-ctx.Done()
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	/*
		fullBody, err := c.GetReqBodyFull()
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to read request body: %v", err))
			return
		}
		reqBodyReader := bytes.NewBuffer(fullBody)
	*/

	reqBodyReader, err := c.GetReqBodyBuffer()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

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
		rb := gitclient.NewRequestBuilder(c.Request.Method, u)
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

		resp, err = gitclient.Do(req)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
			return
		}
		defer resp.Body.Close()
	} else {
		rb := client.NewRequestBuilder(c.Request.Method, u)
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
		defer resp.Body.Close()
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		sizelimit := cfg.Server.SizeLimit * 1024 * 1024
		if err != nil {
			c.Warnf("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, err)
		}
		if err == nil && size > sizelimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			c.Warnf("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, finalURL, size)
			return
		}
	}

	/*
		for key, values := range resp.Header {
			for _, value := range values {
				c.Response.Header.Add(key, value)
			}
		}
	*/
	//copyHeader( resp.Header)
	c.SetHeaders(resp.Header)

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
		c.SetHeader("Cache-Control", "no-store, no-cache, must-revalidate")
		c.SetHeader("Pragma", "no-cache")
		c.SetHeader("Expires", "0")
	}

	bodyReader := resp.Body

	// 读取body内容
	//bodyContent, _ := io.ReadAll(bodyReader)
	//	c.Infof("%s", bodyContent)

	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	c.SetBodyStream(bodyReader, -1)
}
