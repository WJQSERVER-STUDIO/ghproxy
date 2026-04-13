package proxy

import (
	"net/http"
	"testing"

	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

var benchmarkHeaderSource = http.Header{
	"Accept":            {"text/plain"},
	"Accept-Encoding":   {"gzip"},
	"Connection":        {"keep-alive"},
	"User-Agent":        {"curl/8.0.1"},
	"X-Test":            {"one", "two"},
	"CF-Connecting-IP":  {"127.0.0.1"},
	"X-Forwarded-For":   {"127.0.0.1"},
	"Transfer-Encoding": {"chunked"},
}

func BenchmarkMatcherGithubRelease(b *testing.B) {
	cfg := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "header", ForceAllowApi: false},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Matcher("https://github.com/WJQSERVER-STUDIO/ghproxy/releases/download/v1.0.0/asset.tar.gz", cfg)
	}
}

func BenchmarkMatcherRaw(b *testing.B) {
	cfg := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "header", ForceAllowApi: false},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Matcher("https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/README.md", cfg)
	}
}

func BenchmarkMatcherGist(b *testing.B) {
	cfg := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "header", ForceAllowApi: false},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Matcher("https://gist.githubusercontent.com/user/abcdef1234567890/raw/file.txt", cfg)
	}
}

func BenchmarkMatcherAPI(b *testing.B) {
	cfg := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "header", ForceAllowApi: false},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Matcher("https://api.github.com/repos/WJQSERVER-STUDIO/ghproxy/releases", cfg)
	}
}

func BenchmarkSetRequestHeadersClone(b *testing.B) {
	ctx := &touka.Context{Request: &http.Request{Header: benchmarkHeaderSource}}
	cfg := &config.Config{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := &http.Request{Header: make(http.Header)}
		setRequestHeaders(ctx, req, cfg, "clone")
	}
}

func BenchmarkSetRequestHeadersRawCustom(b *testing.B) {
	ctx := &touka.Context{Request: &http.Request{Header: benchmarkHeaderSource}}
	cfg := &config.Config{}
	cfg.Httpc.UseCustomRawHeaders = true

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := &http.Request{Header: make(http.Header)}
		setRequestHeaders(ctx, req, cfg, "raw")
	}
}
