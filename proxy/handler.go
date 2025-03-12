package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"ghproxy/rate"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var exps = []*regexp.Regexp{
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*`),     // 匹配 GitHub Releases 或 Archive 链接
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*`),             // 匹配 GitHub Blob 或 Raw 链接
	regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*`),             // 匹配 GitHub Info 或 Git 相关链接 (例如 .gitattributes, .gitignore)
	regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+`), // 匹配 raw.githubusercontent.com 链接
	regexp.MustCompile(`^(?:https?://)?gist\.github(?:usercontent|)\.com/([^/]+)/.+?/.+`),        // 匹配 gist.githubusercontent.com 链接
	regexp.MustCompile(`^(?:https?://)?api\.github\.com/repos/([^/]+)/([^/]+)/.*`),               // 匹配 api.github.com/repos 链接 (GitHub API)
}

// NoRouteHandler 是 Gin 框架的 NoRoute 处理器函数，用于处理所有未匹配到预定义路由的请求
// 此函数实现了请求的频率限制、URL 路径解析、白名单/黑名单检查、URL 类型匹配和最终的代理请求处理
func NoRouteHandler(cfg *config.Config, limiter *rate.RateLimiter, iplimiter *rate.IPRateLimiter, runMode string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// **频率限制处理**
		if cfg.RateLimit.Enabled { // 检查是否启用频率限制

			var allowed bool // 用于标记是否允许请求

			switch cfg.RateLimit.RateMethod { // 根据配置的频率限制方法选择
			case "ip": // 基于 IP 地址的频率限制
				allowed = iplimiter.Allow(c.ClientIP()) // 使用 IPRateLimiter 检查客户端 IP 是否允许请求
			case "total": // 基于总请求量的频率限制
				allowed = limiter.Allow() // 使用 RateLimiter 检查总请求量是否允许请求
			default: // 无效的频率限制方法
				logWarning("Invalid RateLimit Method") // 记录警告日志
				return                                 // 中断请求处理
			}

			if !allowed { // 如果请求被频率限制阻止
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too Many Requests"})                                                                                           // 返回 429 状态码和错误信息
				logWarning("%s %s %s %s %s 429-TooManyRequests", c.ClientIP(), c.Request.Method, c.Request.URL.RequestURI(), c.Request.Header.Get("User-Agent"), c.Request.Proto) // 记录警告日志
				return                                                                                                                                                            // 中断请求处理
			}
		}

		rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/") // 去掉 URL 前缀的斜杠 '/', 获取原始路径 (例如: /https://github.com/user/repo -> https://github.com/user/repo)
		re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)           // 定义正则表达式，匹配以 http:// 或 https:// 开头的路径，并捕获协议和剩余部分
		matches := re.FindStringSubmatch(rawPath)                      // 使用正则表达式匹配原始路径

		// **路径匹配错误处理**
		if len(matches) < 3 { // 如果匹配结果少于 3 个子串 (完整匹配 + 协议 + 剩余部分)，则说明 URL 格式无效
			errMsg := fmt.Sprintf("%s %s %s %s %s Invalid URL", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto) // 构建错误日志信息
			logWarning(errMsg)                                                                                                                                // 记录警告日志
			c.String(http.StatusForbidden, "Invalid URL Format. Path: %s", rawPath)                                                                           // 返回 403 状态码和错误信息，提示 URL 格式无效
			return                                                                                                                                            // 中断请求处理
		}

		// **构建完整的 URL**
		rawPath = "https://" + matches[2] // 从匹配结果中提取 URL 的剩余部分，并添加 https:// 协议头，构建完整的 URL

		username, repo := MatchUserRepo(rawPath, cfg, c, matches) // 调用 MatchUserRepo 函数，从 URL 中提取用户名和仓库名

		logInfo("%s %s %s %s %s Matched-Username: %s, Matched-Repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, username, repo) // 记录 info 日志，包含匹配到的用户名和仓库名
		// dump log 记录详细信息 c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, full Header
		LogDump("%s %s %s %s %s %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, c.Request.Header) // 记录 dump 日志，包含更详细的请求头信息
		repouser := fmt.Sprintf("%s/%s", username, repo)                                                                                             // 构建 "用户名/仓库名" 格式的字符串

		// **白名单检查**
		if cfg.Whitelist.Enabled { // 检查是否启用白名单
			whitelist := auth.CheckWhitelist(username, repo) // 调用 CheckWhitelist 函数检查当前仓库是否在白名单中
			if !whitelist {                                  // 如果仓库不在白名单中
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Whitelist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser) // 构建错误日志信息
				errMsg := fmt.Sprintf("Whitelist Blocked repo: %s", repouser)                                                                                                                 // 构建返回给客户端的错误信息
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})                                                                                                                          // 返回 403 状态码和 JSON 错误信息
				logWarning(logErrMsg)                                                                                                                                                         // 记录警告日志
				return                                                                                                                                                                        // 中断请求处理
			}
		}

		// **黑名单检查**
		if cfg.Blacklist.Enabled { // 检查是否启用黑名单
			blacklist := auth.CheckBlacklist(username, repo) // 调用 CheckBlacklist 函数检查当前仓库是否在黑名单中
			if blacklist {                                   // 如果仓库在黑名单中
				logErrMsg := fmt.Sprintf("%s %s %s %s %s Blacklist Blocked repo: %s", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, repouser) // 构建错误日志信息
				errMsg := fmt.Sprintf("Blacklist Blocked repo: %s", repouser)                                                                                                                 // 构建返回给客户端的错误信息
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})                                                                                                                          // 返回 403 状态码和 JSON 错误信息
				logWarning(logErrMsg)                                                                                                                                                         // 记录警告日志
				return                                                                                                                                                                        // 中断请求处理
			}
		}

		var matchedIndex = -1 // 用于存储匹配到的正则表达式索引，初始化为 -1 表示未匹配

		// **优化的 URL 匹配逻辑：基于关键词分类匹配**
		switch {
		case strings.Contains(rawPath, "/releases/") || strings.Contains(rawPath, "/archive/"): // 检查 URL 中是否包含 "/releases/" 或 "/archive/" 关键词
			matchedIndex = 0 // 如果包含，则匹配 exps[0] (GitHub Releases/Archive 链接)
		case strings.Contains(rawPath, "/blob/") || strings.Contains(rawPath, "/raw/"): // 检查 URL 中是否包含 "/blob/" 或 "/raw/" 关键词
			matchedIndex = 1 // 如果包含，则匹配 exps[1] (GitHub Blob/Raw 链接)
		case strings.Contains(rawPath, "/info/") || strings.Contains(rawPath, "/git-"): // 检查 URL 中是否包含 "/info/" 或 "/git-" 关键词
			matchedIndex = 2 // 如果包含，则匹配 exps[2] (GitHub Info/Git 相关链接)
		case strings.Contains(rawPath, "raw.githubusercontent.com"): // 检查 URL 中是否包含 "raw.githubusercontent.com" 域名
			matchedIndex = 3 // 如果包含，则匹配 exps[3] (raw.githubusercontent.com 链接)
		case strings.Contains(rawPath, "gist.githubusercontent.com"): // 检查 URL 中是否包含 "gist.githubusercontent.com" 域名
			matchedIndex = 4 // 如果包含，则匹配 exps[4] (gist.githubusercontent.com 链接)
		case strings.Contains(rawPath, "api.github.com/repos/"): // 检查 URL 中是否包含 "api.github.com/repos/" 路径前缀
			matchedIndex = 5 // 如果包含，则匹配 exps[5] (api.github.com/repos 链接)
		}

		if matchedIndex == -1 { // 如果没有任何关键词匹配到，则说明 URL 类型无法识别
			c.AbortWithStatus(http.StatusNotFound)                                                                                                 // 返回 404 状态码
			logWarning("%s %s %s %s %s 404-NOMATCH", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto) // 记录警告日志
			return                                                                                                                                 // 中断请求处理
		}

		// **使用分类匹配到的正则表达式进行精确匹配**
		exp := exps[matchedIndex]
		matches = exp.FindStringSubmatch(rawPath)
		if len(matches) == 0 {
			// 如果精确匹配失败 (例如，关键词匹配到 releases，但实际 URL 格式不符合 releases 的正则)
			c.AbortWithStatus(http.StatusNotFound)
			logWarning("%s %s %s %s %s 404-NOMATCH-ExpSpecific", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto) // 记录警告日志，表明是特定正则匹配失败
			return
		}

		// **HeaderAuth 鉴权检查 (仅针对 api.github.com/repos 链接)**
		if matchedIndex == 5 { // 如果匹配的是 api.github.com/repos 链接 (对应 exps[5])
			if cfg.Auth.AuthMethod != "header" || !cfg.Auth.Enabled { // 检查是否启用了 HeaderAuth 并且 AuthMethod 配置为 "header"
				c.JSON(http.StatusForbidden, gin.H{"error": "HeaderAuth is not enabled."})                                                                                            // 返回 403 状态码和错误信息，提示 HeaderAuth 未启用
				logError("%s %s %s %s %s HeaderAuth-Error: HeaderAuth is not enabled.", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto) // 记录错误日志
				return                                                                                                                                                                // 中断请求处理
			}
		}

		// **处理 blob/raw 路径**
		if matchedIndex == 1 { // 如果匹配的是 GitHub Blob/Raw 链接 (对应 exps[1])
			rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1) // 将 URL 中的 "/blob/" 替换为 "/raw/"，获取 raw 链接 (用于下载原始文件内容)
		}

		// **通用鉴权处理**
		authcheck, err := auth.AuthHandler(c, cfg) // 调用 AuthHandler 函数进行通用鉴权检查 (例如，基于 Cookie 或 Header 的鉴权)
		if !authcheck {                            // 如果鉴权失败
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})                                                                                     // 返回 401 状态码和 JSON 错误信息，提示未授权
			logWarning("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, err) // 记录警告日志，包含鉴权错误信息
			return                                                                                                                                         // 中断请求处理
		}

		// **Debug 日志记录匹配结果**
		logDebug("%s %s %s %s %s Matches: %v", c.ClientIP(), c.Request.Method, rawPath, c.Request.Header.Get("User-Agent"), c.Request.Proto, matches) // 记录 debug 日志，包含匹配结果信息

		// **根据匹配到的 URL 类型，进行不同的代理请求处理**
		switch matchedIndex {
		case 0, 1, 3, 4: // 如果匹配的是 Releases/Archive, Blob/Raw, raw.githubusercontent.com 或 gist.githubusercontent.com 链接 (对应 exps[0], exps[1], exps[3], exps[4])
			//ProxyRequest(c, rawPath, cfg, "chrome", runMode) // 原始的 ProxyRequest 函数 (可能一次性读取全部响应)
			ChunkedProxyRequest(c, rawPath, cfg, "chrome", runMode) // 使用 ChunkedProxyRequest 函数进行分块代理 (更高效，特别是对于大文件)
		case 2: // 如果匹配的是 Info/Git 相关链接 (对应 exps[2])
			//ProxyRequest(c, rawPath, cfg, "git", runMode) // 原始的 ProxyRequest 函数
			GitReq(c, rawPath, cfg, "git", runMode) // 使用 GitReq 函数处理 Git 相关请求 (针对 .gitattributes, .gitignore 等)
		default: // 如果匹配到其他类型 (理论上不应该发生，因为前面的 matchedIndex == -1 已经处理了未识别类型)
			c.String(http.StatusForbidden, "Invalid input.") // 返回 403 状态码和错误信息，提示无效输入
			fmt.Println("Invalid input.")                    // 打印错误信息到控制台
			return                                           // 中断请求处理
		}
	}
}
