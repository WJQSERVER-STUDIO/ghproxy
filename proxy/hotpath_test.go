package proxy

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

func TestNormalizeProxyPath(t *testing.T) {
	testCases := []struct {
		name        string
		rawPath     string
		expected    string
		expectValid bool
	}{
		{name: "Plain host path", rawPath: "/github.com/owner/repo", expected: "github.com/owner/repo", expectValid: true},
		{name: "HTTPS URL", rawPath: "/https://github.com/owner/repo", expected: "github.com/owner/repo", expectValid: true},
		{name: "HTTP URL", rawPath: "http://github.com/owner/repo", expected: "github.com/owner/repo", expectValid: true},
		{name: "Scheme with single slash", rawPath: "https:/github.com/owner/repo", expected: "github.com/owner/repo", expectValid: true},
		{name: "Extra leading slashes", rawPath: "////github.com/owner/repo", expected: "github.com/owner/repo", expectValid: true},
		{name: "Empty path", rawPath: "", expectValid: false},
		{name: "Slash only", rawPath: "////", expectValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := normalizeProxyPath(tc.rawPath)
			if ok != tc.expectValid {
				t.Fatalf("valid = %v, want %v", ok, tc.expectValid)
			}
			if got != tc.expected {
				t.Fatalf("path = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestCopyHeaderFiltered(t *testing.T) {
	src := http.Header{
		"Accept":          {"text/plain"},
		"Connection":      {"keep-alive"},
		"X-Test":          {"one", "two"},
		"Accept-Encoding": {"gzip"},
	}
	dst := make(http.Header)

	copyHeaderFiltered(dst, src, reqHeadersToRemove)

	if got := dst.Values("Accept"); !reflect.DeepEqual(got, []string{"text/plain"}) {
		t.Fatalf("Accept = %v, want [text/plain]", got)
	}
	if got := dst.Values("X-Test"); !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatalf("X-Test = %v, want [one two]", got)
	}
	if got := dst.Values("Connection"); len(got) != 0 {
		t.Fatalf("Connection should be filtered, got %v", got)
	}
	if got := dst.Values("Accept-Encoding"); len(got) != 0 {
		t.Fatalf("Accept-Encoding should be filtered, got %v", got)
	}
}

func TestCopyHeaderFiltered_AllowsAllWhenDenylistEmpty(t *testing.T) {
	src := http.Header{
		"X-Test": {"one", "two"},
	}
	dst := make(http.Header)

	copyHeaderFiltered(dst, src, nil)

	if got := dst.Values("X-Test"); !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatalf("X-Test = %v, want [one two]", got)
	}
}

func TestBuildProxyPath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		matcher  string
		expected string
	}{
		{
			name:     "Blob path rewrites to raw host",
			path:     "github.com/owner/repo/blob/main/file.go",
			matcher:  "blob",
			expected: "https://raw.githubusercontent.com/owner/repo/main/file.go",
		},
		{
			name:     "Non blob path keeps host",
			path:     "raw.githubusercontent.com/owner/repo/main/file.go",
			matcher:  "raw",
			expected: "https://raw.githubusercontent.com/owner/repo/main/file.go",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildProxyPath(tc.path, tc.matcher); got != tc.expected {
				t.Fatalf("buildProxyPath() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestNoRouteHandler_InvalidURI_ReturnsBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://client.example/", nil)
	req.RequestURI = "/"

	ctx, _ := touka.CreateTestContextWithRequest(recorder, req)
	NoRouteHandler(&config.Config{})(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if body := recorder.Body.String(); body == "" {
		t.Fatal("expected error response body to be written")
	}
}

func TestNoRouteHandler_NormalizesAbsoluteRequestURIForAPI(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://client.example/placeholder", nil)
	req.RequestURI = "/https://api.github.com/repos/WJQSERVER-STUDIO/ghproxy/releases?per_page=1"

	ctx, _ := touka.CreateTestContextWithRequest(recorder, req)
	cfg := &config.Config{}
	NoRouteHandler(cfg)(ctx)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if body := recorder.Body.String(); body == "" {
		t.Fatal("expected error response body to be written")
	}
}

func BenchmarkNormalizeProxyPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = normalizeProxyPath("/https://github.com/WJQSERVER-STUDIO/ghproxy/releases/download/v1.0.0/asset.tar.gz")
	}
}

func BenchmarkCopyHeaderFiltered(b *testing.B) {
	src := http.Header{
		"Accept":            {"text/plain"},
		"Accept-Encoding":   {"gzip"},
		"Connection":        {"keep-alive"},
		"User-Agent":        {"curl/8.0.1"},
		"X-Test":            {"one", "two"},
		"CF-Connecting-IP":  {"127.0.0.1"},
		"X-Forwarded-For":   {"127.0.0.1"},
		"Transfer-Encoding": {"chunked"},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dst := make(http.Header)
		copyHeaderFiltered(dst, src, reqHeadersToRemove)
	}
}
