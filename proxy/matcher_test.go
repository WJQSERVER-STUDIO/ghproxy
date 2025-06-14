package proxy

import (
	"ghproxy/config"
	"net/url"
	"reflect"
	"testing"
)

func TestMatcher_Compatibility(t *testing.T) {
	// --- 准备各种配置用于测试 ---
	cfgWithAuth := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "header", ForceAllowApi: false},
	}
	cfgNoAuth := &config.Config{
		Auth: config.AuthConfig{Enabled: false},
	}
	cfgApiForceAllowed := &config.Config{
		Auth: config.AuthConfig{ForceAllowApi: true},
	}
	cfgWrongAuthMethod := &config.Config{
		Auth: config.AuthConfig{Enabled: true, Method: "none"},
	}

	testCases := []struct {
		name            string
		rawPath         string
		config          *config.Config
		expectedUser    string
		expectedRepo    string
		expectedMatcher string
		expectError     bool
		expectedErrCode int
	}{
		{
			name:         "GH Releases Path",
			rawPath:      "https://github.com/owner/repo/releases/download/v1.0/asset.zip",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "releases",
		},
		{
			name:         "GH Archive Path",
			rawPath:      "https://github.com/owner/repo.git/archive/main.zip",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo.git", expectedMatcher: "releases",
		},
		{
			name:         "GH Blob Path",
			rawPath:      "https://github.com/owner/repo/blob/main/path/to/file.go",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "blob",
		},
		{
			name:         "GH Raw Path",
			rawPath:      "https://github.com/owner/repo/raw/main/image.png",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "raw",
		},
		{
			name:         "GH Clone Info Refs",
			rawPath:      "https://github.com/owner/repo.git/info/refs?service=git-upload-pack",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo.git", expectedMatcher: "clone",
		},
		{
			name:         "GH Clone Git Upload Pack",
			rawPath:      "https://github.com/owner/repo/git-upload-pack",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "clone",
		},
		{
			name:        "Girhub Broken Path",
			rawPath:     "https://github.com/owner",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},

		{
			name:         "RawGHUserContent Path",
			rawPath:      "https://raw.githubusercontent.com/owner/repo/branch/file.sh",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "raw",
		},
		{
			name:         "Gist Path",
			rawPath:      "https://gist.github.com/user/abcdef1234567890",
			config:       cfgWithAuth,
			expectedUser: "user", expectedRepo: "", expectedMatcher: "gist",
		},
		{
			name:         "Gist UserContent Path",
			rawPath:      "https://gist.githubusercontent.com/user/abcdef1234567890",
			config:       cfgWithAuth,
			expectedUser: "user", expectedRepo: "", expectedMatcher: "gist",
		},
		{
			name:         "API Repos Path (with Auth)",
			rawPath:      "https://api.github.com/repos/owner/repo/pulls",
			config:       cfgWithAuth,
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "api",
		},
		{
			name:         "API Users Path (with Auth)",
			rawPath:      "https://api.github.com/users/someuser/repos",
			config:       cfgWithAuth,
			expectedUser: "someuser", expectedRepo: "", expectedMatcher: "api",
		},
		{
			name:         "API Other Path (with Auth)",
			rawPath:      "https://api.github.com/octocat",
			config:       cfgWithAuth,
			expectedUser: "", expectedRepo: "", expectedMatcher: "api",
		},
		{
			name:         "API Path (Force Allowed)",
			rawPath:      "https://api.github.com/repos/owner/repo",
			config:       cfgApiForceAllowed, // Auth disabled, but force allowed
			expectedUser: "owner", expectedRepo: "repo", expectedMatcher: "api",
		},
		{
			name:        "Malformed GH Path (no repo)",
			rawPath:     "https://github.com/owner/",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "Malformed GH Path (no action)",
			rawPath:     "https://github.com/owner/repo",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "Malformed GH Path (empty user)",
			rawPath:     "https://github.com//repo/blob/main/file.go",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "Malformed Raw Path (no repo)",
			rawPath:     "https://raw.githubusercontent.com/owner/",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "Malformed Gist Path (no user)",
			rawPath:     "https://gist.github.com/",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "Unsupported GH Action",
			rawPath:     "https://github.com/owner/repo/issues/123",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 400,
		},
		{
			name:        "API Path (No Auth)",
			rawPath:     "https://api.github.com/user",
			config:      cfgNoAuth,
			expectError: true, expectedErrCode: 403,
		},
		{
			name:        "API Path (Wrong Auth Method)",
			rawPath:     "https://api.github.com/user",
			config:      cfgWrongAuthMethod,
			expectError: true, expectedErrCode: 403,
		},
		{
			name:        "No Matcher Found (other domain)",
			rawPath:     "https://bitbucket.org/owner/repo",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 404,
		},
		{
			name:        "No Matcher Found (path too short)",
			rawPath:     "https://a.co",
			config:      cfgWithAuth,
			expectError: true, expectedErrCode: 404,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, repo, matcher, ghErr := Matcher(tc.rawPath, tc.config)

			if tc.expectError {
				if ghErr == nil {
					t.Fatalf("Expected a GHProxyErrors error, but got nil")
				}
				if ghErr.StatusCode != tc.expectedErrCode {
					t.Errorf("Expected error code %d, but got %d (msg: %s)",
						tc.expectedErrCode, ghErr.StatusCode, ghErr.ErrorMessage)
				}
			} else {
				if ghErr != nil {
					t.Fatalf("Expected no error, but got: %s", ghErr.ErrorMessage)
				}
				if user != tc.expectedUser {
					t.Errorf("user: got %q, want %q", user, tc.expectedUser)
				}
				if repo != tc.expectedRepo {
					t.Errorf("repo: got %q, want %q", repo, tc.expectedRepo)
				}
				if matcher != tc.expectedMatcher {
					t.Errorf("matcher: got %q, want %q", matcher, tc.expectedMatcher)
				}
			}
		})
	}
}

func TestExtractParts_Compatibility(t *testing.T) {
	testCases := []struct {
		name          string
		rawURL        string
		expectedOwner string
		expectedRepo  string
		expectedRem   string
		expectedQuery url.Values
		expectError   bool
	}{
		{
			name:          "Standard git clone URL",
			rawURL:        "https://github.com/WJQSERVER-STUDIO/go-utils.git/info/refs?service=git-upload-pack",
			expectedOwner: "/WJQSERVER-STUDIO",
			expectedRepo:  "/go-utils.git",
			expectedRem:   "/info/refs",
			expectedQuery: url.Values{"service": []string{"git-upload-pack"}},
		},
		{
			name:          "No remaining path",
			rawURL:        "https://example.com/owner/repo",
			expectedOwner: "/owner",
			expectedRepo:  "/repo",
			expectedRem:   "",
			expectedQuery: url.Values{},
		},
		{
			name:        "Root path only",
			rawURL:      "https://example.com/",
			expectError: true, // Path is too short
		},
		{
			name:        "One level path",
			rawURL:      "https://example.com/owner",
			expectError: true, // Path is too short
		},
		{
			name:          "Empty path segments",
			rawURL:        "https://example.com//repo/a", // Will be treated as /repo/a
			expectedOwner: "",                            // First part is empty
			expectedRepo:  "/repo",
			expectedRem:   "/a",
		},
		{
			name:        "Invalid URL format",
			rawURL:      "://invalid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, rem, query, err := extractParts(tc.rawURL)

			if (err != nil) != tc.expectError {
				t.Fatalf("extractParts() error = %v, expectError %v", err, tc.expectError)
			}

			if !tc.expectError {
				if owner != tc.expectedOwner {
					t.Errorf("owner: got %q, want %q", owner, tc.expectedOwner)
				}
				if repo != tc.expectedRepo {
					t.Errorf("repo: got %q, want %q", repo, tc.expectedRepo)
				}
				if rem != tc.expectedRem {
					t.Errorf("remaining path: got %q, want %q", rem, tc.expectedRem)
				}
				if !reflect.DeepEqual(query, tc.expectedQuery) {
					t.Errorf("query: got %v, want %v", query, tc.expectedQuery)
				}
			}
		})
	}
}

func TestMatchString_Compatibility(t *testing.T) {
	testCases := []struct {
		target   string
		expected bool
	}{
		{"blob", true}, {"raw", true}, {"gist", true},
		{"clone", false}, {"releases", false},
	}
	for _, tc := range testCases {
		t.Run(tc.target, func(t *testing.T) {
			if got := matchString(tc.target); got != tc.expected {
				t.Errorf("matchString('%s') = %v; want %v", tc.target, got, tc.expected)
			}
		})
	}
}

func BenchmarkMatcher(b *testing.B) {
	cfg := &config.Config{}
	path := "https://github.com/WJQSERVER/speedtest-ex/releases/download/v1.2.0/speedtest-linux-amd64.tar.gz"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = Matcher(path, cfg)
	}
}
