package proxy

/*
func ProxyRequest(c *gin.Context, u string, cfg *config.Config, mode string, runMode string) {
	method := c.Request.Method
	logInfo("%s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto)

	client := createHTTPClient(mode)
	if runMode == "dev" {
		client.DevMode()
	}

	// 发送HEAD请求, 预获取Content-Length
	headReq := client.R()
	setRequestHeaders(c, headReq)
	AuthPassThrough(c, cfg, headReq)

	headResp, err := headReq.Head(u)
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

	req := client.R().SetBody(body)
	setRequestHeaders(c, req)
	AuthPassThrough(c, cfg, req)

	resp, err := SendRequest(c, req, method, u)
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
	if err := copyResponseBody(c, resp.Body); err != nil {
		logError("%s %s %s %s %s Response-Copy-Error: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Proto, err)
	}
}

// 复制响应体
func copyResponseBody(c *gin.Context, respBody io.Reader) error {
	_, err := io.Copy(c.Writer, respBody)
	return err
}

// 判断并选择TLS指纹
func createHTTPClient(mode string) *req.Client {
	client := req.C()
	switch mode {
	case "chrome":
		client.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome()
	case "git":
		client.SetUserAgent("git/2.33.1")
	}
	return client
}

*/
