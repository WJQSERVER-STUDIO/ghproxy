package proxy

import (
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
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

func InitReq(cfgBufferSize int) {
	initChunkedHTTPClient()
	initGitHTTPClient()

	// 初始化固定大小的缓存池
	BufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
}

func initChunkedHTTPClient() {
	ctr = &http.Transport{
		MaxIdleConns:    100,
		MaxConnsPerHost: 60,
		IdleConnTimeout: 20 * time.Second,
	}
	cclient = &http.Client{
		Transport: ctr,
	}
}

func ChunkedProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	// 发送HEAD请求, 预获取Content-Length
	headReq, err := http.NewRequest("HEAD", u, nil)
	if err != nil {
		HandleError(c, fmt.Sprintf("创建HEAD请求失败: %v", err))
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
	defer headResp.Body.Close()

	if err := HandleResponseSize(headResp, cfg, c); err != nil {
		logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
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

	req.Header.Set("Transfer-Encoding", "chunked") // 确保设置分块传输编码
	setRequestHeaders(c, req)
	removeWSHeader(req) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, req)

	resp, err := cclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("发送请求失败: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := HandleResponseSize(resp, cfg, c); err != nil {
		logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	}

	CopyResponseHeaders(resp, c, cfg)

	c.Status(resp.StatusCode)

	// 使用固定32KB缓冲池
	buffer := BufferPool.Get().([]byte)
	defer BufferPool.Put(buffer)

	_, err = io.CopyBuffer(c.Writer, resp.Body, buffer)
	if err != nil {
		logError("%s %s %s %s %s 响应复制错误: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	} else {
		c.Writer.Flush() // 确保刷入
	}
}
