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
)

var (
	gclient *http.Client
	gtr     *http.Transport
)

func initGitHTTPClient() {
	gtr = &http.Transport{
		MaxIdleConns:    30,
		MaxConnsPerHost: 30,
		IdleConnTimeout: 30 * time.Second,
	}
	gclient = &http.Client{
		Transport: gtr,
	}
}

func GitReq(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	// 创建HTTP客户端
	//client := &http.Client{}

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := http.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("创建HEAD请求失败: %v", err))
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
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		HandleError(c, fmt.Sprintf("创建请求失败: %v", err))
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

	/*
		if err := HandleResponseSize(resp, cfg, c); err != nil {
			logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		}
	*/
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
