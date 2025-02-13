package proxy

import (
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var BufferSize int = 32 * 1024 // 32KB

var (
	cclient    *http.Client
	ctr        *http.Transport
	BufferPool *sync.Pool
)

func InitReq(cfg *config.Config) {
	initChunkedHTTPClient(cfg)
	initGitHTTPClient(cfg)

	// 初始化固定大小的缓存池
	BufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
}

func initChunkedHTTPClient(cfg *config.Config) {
	ctr = &http.Transport{
		MaxIdleConns:          100,
		MaxConnsPerHost:       60,
		IdleConnTimeout:       20 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	if cfg.Outbound.Enabled {
		initTransport(cfg, ctr)
	}
	cclient = &http.Client{
		Transport: ctr,
	}
}

func ChunkedProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := http.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	setRequestHeaders(c, headReq)
	removeWSHeader(headReq) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, headReq)

	headResp, err := cclient.Do(headReq)
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

	/*
		if err := HandleResponseSize(headResp, cfg, c); err != nil {
			logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		}
	*/

	body, err := readRequestBody(c)
	if err != nil {
		HandleError(c, err.Error())
		return
	}

	bodyReader := bytes.NewBuffer(body)

	// 创建请求
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	setRequestHeaders(c, req)
	removeWSHeader(req) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, req)

	resp, err := cclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	/*
		if err := HandleResponseSize(resp, cfg, c); err != nil {
			logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
			return
		}
	*/

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

	if cfg.CORS.Enabled {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", "")
	}

	c.Status(resp.StatusCode)

	// 使用固定32KB缓冲池
	buffer := BufferPool.Get().([]byte)
	defer BufferPool.Put(buffer)

	_, err = io.CopyBuffer(c.Writer, resp.Body, buffer)
	if err != nil {
		logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	} else {
		c.Writer.Flush() // 确保刷入
	}
}
