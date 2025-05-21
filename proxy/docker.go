package proxy

import (
	"context"
	"encoding/json"
	"fmt"

	"ghproxy/config"
	"ghproxy/weakcache"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"github.com/cloudwego/hertz/pkg/app"
)

var (
	dockerhubTarget = "registry-1.docker.io"
	ghcrTarget      = "ghcr.io"
)

var cache *weakcache.Cache[string]

type imageInfo struct {
	User  string
	Repo  string
	Image string
}

func InitWeakCache() *weakcache.Cache[string] {
	cache = weakcache.NewCache[string](weakcache.DefaultExpiration, 100)
	return cache
}

/*
func GhcrRouting(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		charToFind := '.'
		reqTarget := c.Param("target")
		path := ""
		target := ""

		if strings.ContainsRune(reqTarget, charToFind) {

			path = c.Param("filepath")
			if reqTarget == "docker.io" {
				target = dockerhubTarget
			} else if reqTarget == "ghcr.io" {
				target = ghcrTarget
			} else {
				target = reqTarget
			}
		} else {
			path = string(c.Request.RequestURI())
		}

		GhcrToTarget(ctx, c, cfg, target, path, nil)

	}
}
*/

func GhcrWithImageRouting(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		charToFind := '.'
		reqTarget := c.Param("target")
		reqImageUser := c.Param("user")
		reqImageName := c.Param("repo")
		reqFilePath := c.Param("filepath")

		path := fmt.Sprintf("%s/%s/%s", reqImageUser, reqImageName, reqFilePath)
		target := ""

		if strings.ContainsRune(reqTarget, charToFind) {

			if reqTarget == "docker.io" {
				target = dockerhubTarget
			} else if reqTarget == "ghcr.io" {
				target = ghcrTarget
			} else {
				target = reqTarget
			}
		} else {
			path = string(c.Request.RequestURI())
			reqImageUser = c.Param("target")
			reqImageName = c.Param("user")
		}
		image := &imageInfo{
			User:  reqImageUser,
			Repo:  reqImageName,
			Image: fmt.Sprintf("%s/%s", reqImageUser, reqImageName),
		}

		GhcrToTarget(ctx, c, cfg, target, path, image)

	}

}

func GhcrToTarget(ctx context.Context, c *app.RequestContext, cfg *config.Config, target string, path string, image *imageInfo) {
	if cfg.Docker.Enabled {
		if target != "" {
			GhcrRequest(ctx, c, "https://"+target+"/v2/"+path+"?"+string(c.Request.QueryString()), image, cfg, target)
		} else {
			if cfg.Docker.Target == "ghcr" {
				GhcrRequest(ctx, c, "https://"+ghcrTarget+string(c.Request.RequestURI()), image, cfg, ghcrTarget)
			} else if cfg.Docker.Target == "dockerhub" {
				GhcrRequest(ctx, c, "https://"+dockerhubTarget+string(c.Request.RequestURI()), image, cfg, dockerhubTarget)
			} else if cfg.Docker.Target != "" {
				// 自定义taget
				GhcrRequest(ctx, c, "https://"+cfg.Docker.Target+string(c.Request.RequestURI()), image, cfg, cfg.Docker.Target)
			} else {
				// 配置为空
				ErrorPage(c, NewErrorWithStatusLookup(403, "Docker Target is not set"))
				return
			}
		}

	} else {
		ErrorPage(c, NewErrorWithStatusLookup(403, "Docker is not Allowed"))
		return
	}
}

func GhcrRequest(ctx context.Context, c *app.RequestContext, u string, image *imageInfo, cfg *config.Config, target string) {

	var (
		method []byte
		req    *http.Request
		resp   *http.Response
		err    error
	)

	go func() {
		<-ctx.Done()
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if req != nil {
			req.Body.Close()
		}
	}()

	method = c.Request.Method()

	rb := ghcrclient.NewRequestBuilder(string(method), u)
	rb.NoDefaultHeaders()
	rb.SetBody(c.Request.BodyStream())
	rb.WithContext(ctx)

	req, err = rb.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	c.Request.Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		headerValue := string(value)
		req.Header.Add(headerKey, headerValue)
	})

	req.Header.Set("Host", target)
	if image != nil {
		token, exist := cache.Get(image.Image)
		if exist {
			logDebug("Use Cache Token: %s", token)
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err = ghcrclient.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}

	// 处理状态码
	if resp.StatusCode == 401 {
		// 请求target /v2/路径
		if string(c.Request.URI().Path()) != "/v2/" {
			resp.Body.Close()
			if image == nil {
				ErrorPage(c, NewErrorWithStatusLookup(401, "Unauthorized"))
				return
			}
			token := ChallengeReq(target, image, ctx, c)

			// 更新kv
			if token != "" {
				logDump("Update Cache Token: %s", token)
				cache.Put(image.Image, token)
			}

			rb := ghcrclient.NewRequestBuilder(string(method), u)
			rb.NoDefaultHeaders()
			rb.SetBody(c.Request.BodyStream())
			rb.WithContext(ctx)

			req, err = rb.Build()
			if err != nil {
				HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
				return
			}

			c.Request.Header.VisitAll(func(key, value []byte) {
				headerKey := string(key)
				headerValue := string(value)
				req.Header.Add(headerKey, headerValue)
			})

			req.Header.Set("Host", target)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err = ghcrclient.Do(req)
			if err != nil {
				HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
				return
			}
		}

	} else if resp.StatusCode == 404 { // 错误处理(404)
		ErrorPage(c, NewErrorWithStatusLookup(404, "Page Not Found (From Github)"))
		return
	}

	var (
		bodySize      int
		contentLength string
		sizelimit     int
	)

	sizelimit = cfg.Server.SizeLimit * 1024 * 1024
	contentLength = resp.Header.Get("Content-Length")
	if contentLength != "" {
		var err error
		bodySize, err = strconv.Atoi(contentLength)
		if err != nil {
			logWarning("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), err)
			bodySize = -1
		}
		if err == nil && bodySize > sizelimit {
			finalURL := resp.Request.URL.String()
			err = resp.Body.Close()
			if err != nil {
				logError("Failed to close response body: %v", err)
			}
			c.Redirect(301, []byte(finalURL))
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), finalURL, bodySize)
			return
		}
	}

	// 复制响应头，排除需要移除的 header
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response.Header.Add(key, value)
		}
	}

	c.Status(resp.StatusCode)

	bodyReader := resp.Body

	if cfg.RateLimit.BandwidthLimit.Enabled {
		bodyReader = limitreader.NewRateLimitedReader(bodyReader, bandwidthLimit, int(bandwidthBurst), ctx)
	}

	if contentLength != "" {
		c.SetBodyStream(bodyReader, bodySize)
		return
	}
	c.SetBodyStream(bodyReader, -1)

}

type AuthToken struct {
	Token string `json:"token"`
}

func ChallengeReq(target string, image *imageInfo, ctx context.Context, c *app.RequestContext) (token string) {
	var resp401 *http.Response
	var req401 *http.Request
	var err error

	rb401 := ghcrclient.NewRequestBuilder("GET", "https://"+target+"/v2/")
	rb401.NoDefaultHeaders()
	rb401.WithContext(ctx)
	rb401.AddHeader("User-Agent", "docker/28.1.1 go/go1.23.8 git-commit/01f442b kernel/6.12.25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/28.1.1 ")
	req401, err = rb401.Build()
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	req401.Header.Set("Host", target)

	resp401, err = ghcrclient.Do(req401)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp401.Body.Close()
	bearer, err := parseBearerWWWAuthenticateHeader(resp401.Header.Get("Www-Authenticate"))
	if err != nil {
		logError("Failed to parse Www-Authenticate header: %v", err)
		return
	}

	scope := fmt.Sprintf("repository:%s:pull", image.Image)

	getAuthRB := ghcrclient.NewRequestBuilder("GET", bearer.Realm).
		NoDefaultHeaders().
		WithContext(ctx).
		AddHeader("User-Agent", "docker/28.1.1 go/go1.23.8 git-commit/01f442b kernel/6.12.25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/28.1.1 ").
		SetHeader("Host", bearer.Service).
		AddQueryParam("service", bearer.Service).
		AddQueryParam("scope", scope)

	getAuthReq, err := getAuthRB.Build()
	if err != nil {
		logError("Failed to create request: %v", err)
		return
	}

	authResp, err := ghcrclient.Do(getAuthReq)
	if err != nil {
		logError("Failed to send request: %v", err)
		return
	}

	defer authResp.Body.Close()

	bodyBytes, err := io.ReadAll(authResp.Body)
	if err != nil {
		logError("Failed to read auth response body: %v", err)
		return
	}

	// 解码json
	var authToken AuthToken
	err = json.Unmarshal(bodyBytes, &authToken)
	if err != nil {
		logError("Failed to decode auth response body: %v", err)
		return
	}
	token = authToken.Token

	return token

}
