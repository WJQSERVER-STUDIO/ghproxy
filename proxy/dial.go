/*
	made&PR by @lfhy
	https://github.com/WJQSERVER-STUDIO/ghproxy/pull/46
*/

package proxy

import (
	"ghproxy/config"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// initTransport 初始化 HTTP 传输层的代理设置
func initTransport(cfg *config.Config, transport *http.Transport) {
	// 如果代理功能未启用，直接返回
	if !cfg.Outbound.Enabled {
		return
	}

	// 如果代理 URL 未设置，使用环境变量中的代理配置
	if cfg.Outbound.Url == "" {
		transport.Proxy = http.ProxyFromEnvironment
		logWarning("Outbound proxy is not set, using environment variables")
		return
	}

	// 尝试解析代理 URL
	proxyInfo, err := url.Parse(cfg.Outbound.Url)
	if err != nil {
		// 如果解析失败，记录错误日志并使用环境变量中的代理配置
		logError("Failed to parse outbound proxy URL %v", err)
		transport.Proxy = http.ProxyFromEnvironment
		return
	}

	// 根据代理 URL 的 scheme（协议类型）选择代理类型
	switch strings.ToLower(proxyInfo.Scheme) {
	case "http", "https": // 如果是 HTTP/HTTPS 代理
		transport.Proxy = http.ProxyURL(proxyInfo) // 设置 HTTP(S) 代理
		logInfo("Using HTTP(S) proxy: %s", proxyInfo.Redacted())
	case "socks5": // 如果是 SOCKS5 代理
		// 调用 newProxyDial 创建 SOCKS5 代理拨号器
		proxyDialer := newProxyDial(cfg.Outbound.Url)
		transport.Proxy = nil // 禁用 HTTP Proxy 设置，因为 SOCKS5 不需要 HTTP Proxy

		// 尝试将 Dialer 转换为支持上下文的 ContextDialer
		if contextDialer, ok := proxyDialer.(proxy.ContextDialer); ok {
			transport.DialContext = contextDialer.DialContext
		} else {
			// 如果不支持 ContextDialer，则回退到传统的 Dial 方法
			transport.Dial = proxyDialer.Dial
			logWarning("SOCKS5 dialer does not support ContextDialer, using legacy Dial")
		}
		logInfo("Using SOCKS5 proxy chain: %s", cfg.Outbound.Url)
	default: // 如果代理协议不支持
		logError("Unsupported proxy scheme: %s", proxyInfo.Scheme)
		transport.Proxy = http.ProxyFromEnvironment // 回退到环境变量代理
	}
}

// newProxyDial 创建一个 SOCKS5 代理拨号器
func newProxyDial(proxyUrls string) proxy.Dialer {
	var proxyDialer proxy.Dialer = proxy.Direct // 初始为直接连接，不使用代理

	// 支持多个代理 URL（以逗号分隔）
	for _, proxyUrl := range strings.Split(proxyUrls, ",") {
		proxyUrl = strings.TrimSpace(proxyUrl) // 去除首尾空格
		if proxyUrl == "" {                    // 跳过空的代理 URL
			continue
		}

		// 解析代理 URL
		urlInfo, err := url.Parse(proxyUrl)
		if err != nil {
			// 如果 URL 解析失败，记录错误日志并跳过
			logError("Failed to parse proxy URL %q: %v", proxyUrl, err)
			continue
		}

		// 检查代理协议是否为 SOCKS5
		if urlInfo.Scheme != "socks5" {
			logWarning("Skipping non-SOCKS5 proxy: %s", urlInfo.Scheme)
			continue
		}

		// 解析代理认证信息（用户名和密码）
		auth := parseAuth(urlInfo)

		// 创建 SOCKS5 代理拨号器
		dialer, err := createSocksDialer(urlInfo.Host, auth, proxyDialer)
		if err != nil {
			// 如果创建失败，记录错误日志并跳过
			logError("Failed to create SOCKS5 dialer for %q: %v", proxyUrl, err)
			continue
		}

		// 更新代理拨号器，支持代理链
		proxyDialer = dialer
	}

	return proxyDialer
}

// parseAuth 解析代理 URL 中的认证信息（用户名和密码）
func parseAuth(urlInfo *url.URL) *proxy.Auth {
	// 如果 URL 中没有用户信息，返回 nil
	if urlInfo.User == nil {
		return nil
	}

	// 获取用户名
	username := urlInfo.User.Username()

	// 获取密码（注意：Password() 返回两个值，需要显式处理第二个值）
	password, passwordSet := urlInfo.User.Password()
	if !passwordSet {
		password = "" // 如果密码未设置，使用空字符串
	}

	// 返回包含用户名和密码的认证信息
	return &proxy.Auth{
		User:     username,
		Password: password, // 允许空密码
	}
}

// createSocksDialer 创建 SOCKS5 拨号器
func createSocksDialer(host string, auth *proxy.Auth, previous proxy.Dialer) (proxy.Dialer, error) {
	// 调用 golang.org/x/net/proxy 提供的 SOCKS5 方法创建拨号器
	return proxy.SOCKS5("tcp", host, auth, previous)
}
