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
	setRequestHeaders(c, req)
	AuthPassThrough(c, cfg, req)

	resp, err := gclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	if err := HandleResponseSize(resp, cfg, c); err != nil {
		logWarning("%s %s %s %s %s Response-Size-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	}

	CopyResponseHeaders(resp, c, cfg)
	c.Status(resp.StatusCode)
	if err := gitCopyResponseBody(c, resp.Body); err != nil {
		logError("%s %s %s %s %s Response-Copy-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
	}
}

// 复制响应体
func gitCopyResponseBody(c *gin.Context, respBody io.Reader) error {
	_, err := io.Copy(c.Writer, respBody)
	return err
}
