package proxy

import (
	"ghproxy/config"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

func newProxyDial(prxoyUrls string) proxy.Dialer {
	var proxyDialer proxy.Dialer = proxy.Direct
	for _, prxoyUrl := range strings.Split(prxoyUrls, ",") {
		urlInfo, err := url.Parse(prxoyUrl)
		if err != nil {
			return proxyDialer
		}
		var auth *proxy.Auth = nil
		if urlInfo.User != nil {
			pwd, _ := urlInfo.User.Password()
			auth = &proxy.Auth{
				User:     urlInfo.User.Username(),
				Password: pwd,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", urlInfo.Host, auth, proxyDialer)
		if err == nil {
			proxyDialer = dialer
		}
	}
	return proxyDialer
}

func initTransport(cfg *config.Config, transport *http.Transport) {
	if !cfg.Proxy.Enabled {
		return
	}
	if cfg.Proxy.Url == "" {
		transport.Proxy = http.ProxyFromEnvironment
		return
	}

	proxyInfo, err := url.Parse(cfg.Proxy.Url)
	if err == nil {
		if strings.HasPrefix(cfg.Proxy.Url, "http") {
			transport.Proxy = http.ProxyURL(proxyInfo)
		} else {
			proxyDialer := newProxyDial(cfg.Proxy.Url)
			transport.Dial = proxyDialer.Dial
			transport.DialContext = proxyDialer.(proxy.ContextDialer).DialContext
		}
	}
}
