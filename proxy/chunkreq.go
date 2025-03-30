package proxy

import (
	"bytes"
	"context"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	//hresp "github.com/cloudwego/hertz/pkg/protocol/http1/resp"
)

func ChunkedProxyRequest(ctx context.Context, c *app.RequestContext, u string, cfg *config.Config, matcher string) {
	method := c.Request.Method

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := client.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, headReq)
	removeWSHeader(headReq) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, headReq)

	headResp, err := client.Do(headReq)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	//defer headResp.Body.Close()
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logError("Failed to close response body: %v", err)
		}
	}(headResp.Body)

	contentLength := headResp.Header.Get("Content-Length")
	sizelimit := cfg.Server.SizeLimit * 1024 * 1024
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > sizelimit {
			finalURL := headResp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, []byte(finalURL))
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol(), finalURL, size)
			return
		}
	}

	body := c.Request.Body()

	bodyReader := bytes.NewBuffer(body)

	req, err := client.NewRequest(string(method()), u, bodyReader)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, req)
	removeWSHeader(req) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, req)

	resp, err := client.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	//defer resp.Body.Close()

	// 错误处理(404)
	if resp.StatusCode == 404 {
		c.String(http.StatusNotFound, "File Not Found")
		return
	}

	contentLength = resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		if err == nil && size > sizelimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, []byte(finalURL))
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), finalURL, size)
			return
		}
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
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
	//c.Response.HijackWriter(hresp.NewChunkedBodyWriter(&c.Response, c.GetWriter()))

	if MatcherShell(u) && matchString(matcher, matchedMatchers) && cfg.Shell.Editor {
		// 判断body是不是gzip
		var compress string
		if resp.Header.Get("Content-Encoding") == "gzip" {
			compress = "gzip"
		}

		logInfo("Is Shell: %s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol())
		c.Header("Content-Length", "")

		//err := ProcessLinksAndWriteChunked(resp.Body, compress, string(c.Request.Host()), cfg, c)
		reader, _, err := processLinks(resp.Body, compress, string(c.Request.Host()), cfg)
		c.SetBodyStream(reader, -1)

		if err != nil {
			logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol(), err)
			return
		}
	} else {
		//err = hwriter.Writer(resp.Body, c)
		//writer := c.Response.BodyWriter()

		/*
			_, err := copyb.Copy(c.Response.BodyWriter(), resp.Body)
			if err != nil {
				logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol(), err)
				return
			} else {
				c.Flush() // 确保刷入
			}
		*/
		c.SetBodyStream(resp.Body, -1)
	}

}
