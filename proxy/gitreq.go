package proxy

import (
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	httpc "github.com/satomitouka/touka-httpc"
)

var (
	gclient *httpc.Client
	gtr     *http.Transport
)

func initGitHTTPClient(cfg *config.Config) {
	gtr = &http.Transport{
		MaxIdleConns:    30,
		MaxConnsPerHost: 30,
		IdleConnTimeout: 30 * time.Second,
	}
	if cfg.Outbound.Enabled {
		initTransport(cfg, gtr)
	}
	/*
		gclient = &http.Client{
			Transport: gtr,
		}
	*/
	gclient = httpc.New(
		httpc.WithTransport(gtr),
	)
}

func GitReq(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := gclient.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, headReq)
	AuthPassThrough(c, cfg, headReq)

	headResp, err := gclient.Do(headReq)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}

	// defer headResp.Body.Close()
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

	// 创建请求
	req, err := gclient.NewRequest(method, u, bodyReader)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, req)
	AuthPassThrough(c, cfg, req)

	resp, err := gclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	//defer resp.Body.Close()
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logError("Failed to close response body: %v", err)
		}
	}(resp.Body)

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

	if cfg.CORS.Enabled {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}

	c.Status(resp.StatusCode)

	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		logError("%s %s %s %s %s Response-Copy-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
	}
}
