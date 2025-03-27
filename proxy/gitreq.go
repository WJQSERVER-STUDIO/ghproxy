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

func GitReq(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	logDump("Url Before FMT:%s", u)
	if cfg.GitClone.Mode == "cache" {
		userPath, repoPath, remainingPath, queryParams, err := extractParts(u)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to extract parts from URL: %v", err))
			return
		}
		// 构建新url
		u = cfg.GitClone.SmartGitAddr + userPath + repoPath + remainingPath + "?" + queryParams.Encode()
		logDump("New Url After FMT:%s", u)
	}

	var (
		resp *http.Response
		err  error
	)

	body, err := readRequestBody(c)
	if err != nil {
		HandleError(c, err.Error())
		return
	}

	bodyReader := bytes.NewBuffer(body)
	// 创建请求

	if cfg.GitClone.Mode == "cache" {
		req, err := gitclient.NewRequest(method, u, bodyReader)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
			return
		}
		setRequestHeaders(c, req)
		removeWSHeader(req)
		reWriteEncodeHeader(req)
		AuthPassThrough(c, cfg, req)

		resp, err = gitclient.Do(req)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
			return
		}
	} else {
		req, err := client.NewRequest(method, u, bodyReader)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
			return
		}
		setRequestHeaders(c, req)
		removeWSHeader(req)
		reWriteEncodeHeader(req)
		AuthPassThrough(c, cfg, req)

		resp, err = client.Do(req)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
			return
		}
	}
	//defer resp.Body.Close()
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logError("Failed to close response body: %v", err)
		}
	}(resp.Body)

	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, err := strconv.Atoi(contentLength)
		sizelimit := cfg.Server.SizeLimit * 1024 * 1024
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
	_, err = copyb.Copy(c.Writer, resp.Body)

	if err != nil {
		logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
		return
	} else {

		c.Writer.Flush() // 确保刷入
	}

}
