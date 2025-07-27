package proxy

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ghproxy/config"
	"ghproxy/weakcache"

	"github.com/WJQSERVER-STUDIO/go-utils/iox"
	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/go-json-experiment/json"
	"github.com/infinite-iroha/touka"
)

var (
	dockerhubTarget = "registry-1.docker.io"
	ghcrTarget      = "ghcr.io"
)

// cache 用于存储认证令牌, 避免重复获取
var cache *weakcache.Cache[string]

// imageInfo 结构体用于存储镜像的相关信息
type imageInfo struct {
	User  string
	Repo  string
	Image string
}

// InitWeakCache 初始化弱引用缓存
func InitWeakCache() *weakcache.Cache[string] {
	// 使用默认过期时间和容量为100创建一个新的弱引用缓存
	cache = weakcache.NewCache[string](weakcache.DefaultExpiration, 100)
	return cache
}

// GhcrWithImageRouting 处理带有镜像路由的请求, 根据目标路由到不同的Docker注册表
func GhcrWithImageRouting(cfg *config.Config) touka.HandlerFunc {
	return func(c *touka.Context) {
		// 从 main.go 中固定的路由 "/v2/:target/:user/:repo/*filepath" 获取参数
		reqTarget := c.Param("target")
		reqImageUser := c.Param("user")
		reqImageName := c.Param("repo")
		reqFilePath := c.Param("filepath")

		var upstreamTarget string
		var requestPath string
		var imageNameForAuth string

		// 关键逻辑: 判断 reqTarget 是真实主机名还是镜像名的一部分
		// 依据: 真实主机名/IP通常包含'.'或':'
		if strings.Contains(reqTarget, ".") || strings.Contains(reqTarget, ":") {
			// 情况 A: reqTarget 是一个显式指定的主机名 (例如 "ghcr.io", "my-registry.com", "127.0.0.1:5000")
			c.Debugf("Request target '%s' identified as an explicit hostname.", reqTarget)
			upstreamTarget = reqTarget
			// 上游请求的路径是主机名之后的部分
			requestPath = fmt.Sprintf("%s/%s%s", reqImageUser, reqImageName, reqFilePath)
			// 用于认证的镜像名是 user/repo
			imageNameForAuth = fmt.Sprintf("%s/%s", reqImageUser, reqImageName)
		} else {
			// 情况 B: reqTarget 是镜像名的一部分 (例如 "wjqserver", "library")
			c.Debugf("Request target '%s' identified as part of an image name. Using default registry.", reqTarget)
			// 使用配置文件中的默认目标
			switch cfg.Docker.Target {
			case "ghcr":
				upstreamTarget = ghcrTarget
			case "dockerhub":
				upstreamTarget = dockerhubTarget
			case "":
				ErrorPage(c, NewErrorWithStatusLookup(500, "Default Docker Target is not configured in config file"))
				return
			default:
				upstreamTarget = cfg.Docker.Target
			}
			// 必须将路由错误分割的所有部分重新组合成完整的镜像路径
			requestPath = fmt.Sprintf("%s/%s/%s%s", reqTarget, reqImageUser, reqImageName, reqFilePath)
			// 用于认证的镜像名是 target/user (例如 "wjqserver/ghproxy", "library/ubuntu")
			imageNameForAuth = fmt.Sprintf("%s/%s", reqTarget, reqImageUser)
		}

		// 清理路径, 防止出现 "//"
		requestPath = strings.TrimPrefix(requestPath, "/")

		// 为认证和缓存准备镜像信息
		image := &imageInfo{
			Image: imageNameForAuth,
		}

		// 调用 GhcrToTarget 处理实际的代理请求
		GhcrToTarget(c, cfg, upstreamTarget, requestPath, image)
	}
}

// GhcrToTarget 根据配置和目标信息将请求代理到上游Docker注册表
func GhcrToTarget(c *touka.Context, cfg *config.Config, target string, path string, image *imageInfo) {
	// 检查Docker代理是否启用
	if !cfg.Docker.Enabled {
		ErrorPage(c, NewErrorWithStatusLookup(403, "Docker is not Allowed"))
		return
	}

	var ctx = c.Request.Context()

	// 构造目标URL. 这里的target和path都是由GhcrWithImageRouting正确解析得来的.
	destUrl := "https://" + target + "/v2/" + path
	if query := c.GetReqQueryString(); query != "" {
		destUrl += "?" + query
	}
	c.Debugf("Proxying to target '%s' with path '%s'. Final URL: %s", target, path, destUrl)

	// 执行实际的代理请求
	GhcrRequest(ctx, c, destUrl, image, cfg, target)
}

// GhcrRequest 执行对Docker注册表的HTTP请求, 处理认证和重定向
func GhcrRequest(ctx context.Context, c *touka.Context, u string, image *imageInfo, cfg *config.Config, target string) {
	var (
		method string
		req    *http.Request
		resp   *http.Response
		err    error
	)

	method = c.Request.Method
	ghcrclient := c.GetHTTPC()
	bodyByte, err := c.GetReqBodyFull()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

	// 构建初始请求
	rb := ghcrclient.NewRequestBuilder(method, u)
	rb.NoDefaultHeaders()                 // 不使用默认头部, 以便完全控制
	rb.SetBody(bytes.NewBuffer(bodyByte)) // 设置请求体
	rb.WithContext(ctx)                   // 设置请求上下文

	req, err = rb.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	// 复制客户端请求的头部到代理请求
	copyHeader(c.Request.Header, req.Header)

	// 确保 Accept 头部被正确设置
	if acceptHeader, ok := c.Request.Header["Accept"]; ok {
		req.Header["Accept"] = acceptHeader
	}

	// 设置 Host 头部为上游目标
	req.Header.Set("Host", target)

	// 尝试从缓存中获取并使用认证令牌
	if image != nil && image.Image != "" {
		token, exist := cache.Get(image.Image)
		if exist {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	// 发送初始请求
	resp, err = ghcrclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}

	// 处理 401 Unauthorized 或 404 Not Found 响应, 尝试重新认证并重试
	if resp.StatusCode == 401 || resp.StatusCode == 404 {
		// 对于 /v2/ 的请求不进行重试, 因为它通常用于发现认证端点
		shouldRetry := string(c.GetRequestURIPath()) != "/v2/"
		originalStatusCode := resp.StatusCode
		c.Debugf("Initial request failed with status %d. Retry eligibility: %t", originalStatusCode, shouldRetry)

		if shouldRetry {
			if image == nil || image.Image == "" {
				_ = resp.Body.Close() // 终止流程, 关闭当前响应体
				ErrorPage(c, NewErrorWithStatusLookup(originalStatusCode, "Unauthorized"))
				return
			}
			// 获取新的认证令牌
			token := ChallengeReq(target, image, ctx, c)

			if token != "" {
				c.Debugf("Successfully obtained auth token. Retrying request.")
				_ = resp.Body.Close() // 在发起重试请求前, 关闭旧的响应体

				// 更新kv
				c.Debugf("Update Cache Token: %s", token)
				cache.Put(image.Image, token)

				// 重新构建并发送请求
				rb_retry := ghcrclient.NewRequestBuilder(method, u)
				rb_retry.NoDefaultHeaders()
				rb_retry.SetBody(bytes.NewBuffer(bodyByte))
				rb_retry.WithContext(ctx)

				req_retry, err_retry := rb_retry.Build()
				if err_retry != nil {
					HandleError(c, fmt.Sprintf("Failed to create retry request: %v", err_retry))
					return
				}

				copyHeader(c.Request.Header, req_retry.Header) // 复制原始头部
				if acceptHeader, ok := c.Request.Header["Accept"]; ok {
					req_retry.Header["Accept"] = acceptHeader
				}

				req_retry.Header.Set("Host", target)                   // 设置 Host 头部
				req_retry.Header.Set("Authorization", "Bearer "+token) // 使用新令牌

				c.Debugf("Executing retry request. Method: %s, URL: %s", req_retry.Method, req_retry.URL.String())

				resp_retry, err_retry := ghcrclient.Do(req_retry)
				if err_retry != nil {
					HandleError(c, fmt.Sprintf("Failed to send retry request: %v", err_retry))
					return
				}
				c.Debugf("Retry request completed with status code: %d", resp_retry.StatusCode)
				resp = resp_retry // 更新响应为重试后的响应
			} else {
				c.Warnf("Failed to obtain auth token. Cannot retry.")
				// 获取令牌失败, 将继续处理原始的401/404响应, 其响应体仍然打开
			}
		}
	}

	// 透明地处理 302 Found 或 307 Temporary Redirect 重定向
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
		location := resp.Header.Get("Location")
		if location == "" {
			_ = resp.Body.Close() // 终止流程, 关闭当前响应体
			HandleError(c, "Redirect response missing Location header")
			return
		}

		redirectURL, err := url.Parse(location)
		if err != nil {
			_ = resp.Body.Close() // 终止流程, 关闭当前响应体
			HandleError(c, fmt.Sprintf("Failed to parse redirect location: %v", err))
			return
		}

		// 如果 Location 是相对路径, 则根据原始请求的 URL 解析为绝对路径
		if !redirectURL.IsAbs() {
			originalURL := resp.Request.URL
			redirectURL = originalURL.ResolveReference(redirectURL)
			c.Debugf("Resolved relative redirect to absolute URL: %s", redirectURL.String())
		}

		c.Debugf("Handling redirect. Status: %d, Final Location: %s", resp.StatusCode, redirectURL.String())
		_ = resp.Body.Close() // 明确关闭重定向响应的响应体, 因为我们将发起新请求

		// 创建并发送重定向请求, 通常使用 GET 方法
		redirectReq, err := http.NewRequestWithContext(ctx, "GET", redirectURL.String(), nil)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to create redirect request: %v", err))
			return
		}
		redirectReq.Header.Set("User-Agent", c.Request.UserAgent()) // 复制 User-Agent

		c.Debugf("Executing redirect request to: %s", redirectURL.String())
		redirectResp, err := ghcrclient.Do(redirectReq)
		if err != nil {
			HandleError(c, fmt.Sprintf("Failed to execute redirect request to %s: %v", redirectURL.String(), err))
			return
		}
		c.Debugf("Redirect request to %s completed with status %d", redirectURL.String(), redirectResp.StatusCode)
		resp = redirectResp // 更新响应为重定向后的响应
	}

	// 如果最终响应是 404, 则读取响应体并返回自定义错误页面
	if resp.StatusCode == 404 {
		defer resp.Body.Close() // 使用defer确保在函数返回前关闭响应体
		bodyBytes, err := iox.ReadAll(resp.Body)
		if err != nil {
			c.Warnf("Failed to read upstream 404 response body: %v", err)
		} else {
			c.Warnf("Upstream 404 response body: %s", string(bodyBytes))
		}
		ErrorPage(c, NewErrorWithStatusLookup(404, "Page Not Found (From Upstream)"))
		return
	}

	var (
		bodySize      int
		contentLength string
		sizelimit     int
	)

	// 获取配置中的大小限制并转换单位 (MB -> Byte)
	sizelimit = cfg.Server.SizeLimit * 1024 * 1024
	contentLength = resp.Header.Get("Content-Length")
	if contentLength != "" {
		var err error
		bodySize, err = strconv.Atoi(contentLength)
		if err != nil {
			c.Warnf("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, err)
			bodySize = -1 // 无法解析则设置为 -1
		}
		// 如果内容大小超出限制, 返回 301 重定向到原始上游URL
		if err == nil && bodySize > sizelimit {
			finalURL := resp.Request.URL.String()
			_ = resp.Body.Close() // 明确关闭响应体, 因为我们将重定向而不是流式传输
			c.Redirect(301, finalURL)
			c.Warnf("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, finalURL, bodySize)
			return
		}
	}

	// 将上游响应头部复制到客户端响应
	c.SetHeaders(resp.Header)
	// 设置客户端响应状态码
	c.Status(resp.StatusCode)
	// bodyReader 的所有权将转移给 SetBodyStream, 不再由此函数管理关闭
	bodyReader := resp.Body

	// 如果启用了带宽限制, 则使用限速读取器
	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	// 根据 Content-Length 设置响应体流
	if contentLength != "" {
		c.SetBodyStream(bodyReader, bodySize)
		return
	}
	c.SetBodyStream(bodyReader, -1)
}

// AuthToken 用于解析认证响应中的令牌
type AuthToken struct {
	Token string `json:"token"`
}

// ChallengeReq 执行认证挑战流程, 获取新的认证令牌
func ChallengeReq(target string, image *imageInfo, ctx context.Context, c *touka.Context) (token string) {
	var resp401 *http.Response
	var req401 *http.Request
	var err error
	ghcrclient := c.GetHTTPC()

	// 对 /v2/ 端点发送 GET 请求以触发认证挑战
	rb401 := ghcrclient.NewRequestBuilder("GET", "https://"+target+"/v2/")
	rb401.NoDefaultHeaders()
	rb401.WithContext(ctx)
	req401, err = rb401.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	req401.Header.Set("Host", target) // 设置 Host 头部

	resp401, err = ghcrclient.Do(req401)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp401.Body.Close() // 确保响应体关闭

	// 解析 Www-Authenticate 头部, 获取认证领域和参数
	bearer, err := parseBearerWWWAuthenticateHeader(resp401.Header.Get("Www-Authenticate"))
	if err != nil {
		c.Errorf("Failed to parse Www-Authenticate header: %v", err)
		return
	}

	// 构建认证范围 (scope), 通常是 repository:<image_name>:pull
	scope := fmt.Sprintf("repository:%s:pull", image.Image)

	// 使用解析到的 Realm 和 Service, 以及 scope 请求认证令牌
	getAuthRB := ghcrclient.NewRequestBuilder("GET", bearer.Realm).
		NoDefaultHeaders().
		WithContext(ctx).
		SetHeader("Host", bearer.Service).
		AddQueryParam("service", bearer.Service).
		AddQueryParam("scope", scope)

	getAuthReq, err := getAuthRB.Build()
	if err != nil {
		c.Errorf("Failed to create request: %v", err)
		return
	}

	authResp, err := ghcrclient.Do(getAuthReq)
	if err != nil {
		c.Errorf("Failed to send request: %v", err)
		return
	}
	defer authResp.Body.Close() // 确保响应体关闭

	// 读取认证响应体
	bodyBytes, err := iox.ReadAll(authResp.Body)
	if err != nil {
		c.Errorf("Failed to read auth response body: %v", err)
		return
	}

	// 解码 JSON 响应以获取令牌
	var authToken AuthToken
	err = json.Unmarshal(bodyBytes, &authToken)
	if err != nil {
		c.Errorf("Failed to decode auth response body: %v", err)
		return
	}
	token = authToken.Token // 提取令牌

	return token
}
