package proxy

import (
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var chunkedBufferSize int

var (
	client *http.Client
	tr     *http.Transport
)

func InitChunkedReq(cfgBufferSize int) {
	initChunkedBufferSize(cfgBufferSize)
	initChunkedHTTPClient()
}

func initChunkedBufferSize(cfgBufferSize int) {
	if cfgBufferSize == 0 {
		chunkedBufferSize = 4096 // 默认缓冲区大小
	} else {
		chunkedBufferSize = cfgBufferSize
	}
}

func initChunkedHTTPClient() {
	tr = &http.Transport{
		MaxIdleConns:    100,
		MaxConnsPerHost: 60,
		IdleConnTimeout: 15 * time.Second,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
}

func ChunkedProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
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
	removeWSHeader(headReq) // 删除Conection Upgrade头, 避免与HTTP/2冲突(检查是否存在Upgrade头)
	AuthPassThrough(c, cfg, headReq)

	headResp, err := client.Do(headReq)
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

	resp, err := client.Do(req)
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

	if err := chunkedCopyResponseBody(c, resp.Body); err != nil {
		logError("%s %s %s %s %s 响应复制错误: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
	}
}

// 复制响应体
func chunkedCopyResponseBody(c *gin.Context, respBody io.Reader) error {
	buf := make([]byte, chunkedBufferSize)
	for {
		n, err := respBody.Read(buf)
		if n > 0 {
			if _, err := c.Writer.Write(buf[:n]); err != nil {
				return err
			}
			c.Writer.Flush() // 确保每次写入后刷新
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}
