package proxy

import (
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"strconv"

	"github.com/WJQSERVER-STUDIO/go-utils/copyb"
	"github.com/gin-gonic/gin"
)

func ChunkedProxyRequest(c *gin.Context, u string, cfg *config.Config, matcher string) {
	method := c.Request.Method

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := client.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, headReq)
	removeWSHeader(headReq) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	reWriteEncodeHeader(headReq)
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
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto, finalURL, size)
			return
		}
	}

	body, err := readRequestBody(c)
	if err != nil {
		HandleError(c, err.Error())
		return
	}

	bodyReader := bytes.NewBuffer(body)

	req, err := client.NewRequest(method, u, bodyReader)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, req)
	removeWSHeader(req) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	reWriteEncodeHeader(req)
	AuthPassThrough(c, cfg, req)

	resp, err := client.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

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
			c.Redirect(http.StatusMovedPermanently, finalURL)
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto, finalURL, size)
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

	//c.Header("Accept-Encoding", "gzip")
	//c.Header("Content-Encoding", "gzip")

	/*
		if cfg.CORS.Enabled {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", "")
		}
	*/

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

	if MatcherShell(u) && matchString(matcher, matchedMatchers) && cfg.Shell.Editor {
		// 判断body是不是gzip
		var compress string
		if resp.Header.Get("Content-Encoding") == "gzip" {
			compress = "gzip"
		}

		logInfo("Is Shell: %s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)
		c.Header("Content-Length", "")
		_, err = processLinks(resp.Body, c.Writer, compress, c.Request.Host, cfg)
		if err != nil {
			logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		} else {
			c.Writer.Flush() // 确保刷入
		}
	} else {
		//_, err = io.CopyBuffer(c.Writer, resp.Body, nil)
		_, err = copyb.CopyBuffer(c.Writer, resp.Body, nil)
		if err != nil {
			logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		} else {
			c.Writer.Flush() // 确保刷入
		}
	}
}
