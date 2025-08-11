package proxy

import (
	"fmt"
	"ghproxy/config"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var (
	githubPrefixLen      int
	rawPrefixLen         int
	gistPrefixLen        int
	gistContentPrefixLen int
	apiPrefixLen         int
)

const (
	githubPrefix            = "https://github.com/"
	rawPrefix               = "https://raw.githubusercontent.com/"
	gistPrefix              = "https://gist.github.com/"
	gistContentPrefix       = "https://gist.githubusercontent.com/"
	apiPrefix               = "https://api.github.com/"
	ociv2Prefix             = "https://v2/"
	releasesDownloadSnippet = "releases/download/"
)

func init() {
	githubPrefixLen = len(githubPrefix)
	rawPrefixLen = len(rawPrefix)
	gistPrefixLen = len(gistPrefix)
	gistContentPrefixLen = len(gistContentPrefix)
	apiPrefixLen = len(apiPrefix)
}

// Matcher 从原始URL路径中高效地解析并匹配代理规则.
func Matcher(rawPath string, cfg *config.Config) (string, string, string, *GHProxyErrors) {
	/*
		if len(rawPath) < 18 {
			return "", "", "", NewErrorWithStatusLookup(404, "path too short")
		}
	*/

	// 匹配 "https://github.com/"
	if strings.HasPrefix(rawPath, githubPrefix) {
		pathAfterDomain := rawPath[githubPrefixLen:]

		// 解析 user
		i := strings.IndexByte(pathAfterDomain, '/')
		if i <= 0 {
			return "", "", "", NewErrorWithStatusLookup(400, "malformed github path: missing user")
		}
		user := pathAfterDomain[:i]
		pathAfterUser := pathAfterDomain[i+1:]

		// 解析 repo
		i = strings.IndexByte(pathAfterUser, '/')
		if i <= 0 {
			return "", "", "", NewErrorWithStatusLookup(400, "malformed github path: missing action")
		}
		repo := pathAfterUser[:i]
		pathAfterRepo := pathAfterUser[i+1:]

		if len(pathAfterRepo) == 0 {
			return "", "", "", NewErrorWithStatusLookup(400, "malformed github path: missing action")
		}

		// 优先处理所有 "releases" 相关的下载路径
		if strings.HasPrefix(pathAfterRepo, "releases/") {
			// 情况 A: "releases/download/..."
			if strings.HasPrefix(pathAfterRepo, "releases/download/") {
				return user, repo, "releases", nil
			}
			// 情况 B: "releases/:tag/download/..."
			pathAfterReleases := pathAfterRepo[len("releases/"):]
			slashIndex := strings.IndexByte(pathAfterReleases, '/')
			if slashIndex > 0 { // 确保tag不为空
				pathAfterTag := pathAfterReleases[slashIndex+1:]
				if strings.HasPrefix(pathAfterTag, "download/") {
					return user, repo, "releases", nil
				}
			}
			// 如果不满足上述下载链接的结构, 则为网页浏览路径, 予以拒绝
			return "", "", "", NewErrorWithStatusLookup(400, "unsupported releases page, only download links are allowed")
		}

		// 检查 "archive/" 路径
		if strings.HasPrefix(pathAfterRepo, "archive/") {
			// 根据测试用例, archive路径的matcher也应为releases
			return user, repo, "releases", nil
		}

		// 如果不是下载路径, 则解析action并进行分类
		i = strings.IndexByte(pathAfterRepo, '/')
		action := pathAfterRepo
		if i != -1 {
			action = pathAfterRepo[:i]
		}

		var matcher string
		switch action {
		case "blob":
			matcher = "blob"
		case "raw":
			matcher = "raw"
		case "info", "git-upload-pack":
			matcher = "clone"
		default:
			return "", "", "", NewErrorWithStatusLookup(400, fmt.Sprintf("unsupported github action: %s", action))
		}
		return user, repo, matcher, nil
	}

	// 匹配 "https://raw.githubusercontent.com/"
	if strings.HasPrefix(rawPath, rawPrefix) {
		remaining := rawPath[rawPrefixLen:]
		parts := strings.SplitN(remaining, "/", 3)
		if len(parts) < 3 {
			return "", "", "", NewErrorWithStatusLookup(400, "malformed raw url: path too short")
		}
		return parts[0], parts[1], "raw", nil
	}

	// 匹配 "https://gist.github.com/" 或 "https://gist.githubusercontent.com/"
	isGist := strings.HasPrefix(rawPath, gistPrefix)
	if isGist || strings.HasPrefix(rawPath, gistContentPrefix) {
		var remaining string
		if isGist {
			remaining = rawPath[gistPrefixLen:]
		} else {
			remaining = rawPath[gistContentPrefixLen:]
		}
		parts := strings.SplitN(remaining, "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			return "", "", "", NewErrorWithStatusLookup(400, "malformed gist url: missing user")
		}
		return parts[0], "", "gist", nil
	}

	// 匹配 "https://api.github.com/"
	if strings.HasPrefix(rawPath, apiPrefix) {
		if !cfg.Auth.ForceAllowApi && (cfg.Auth.Method != "header" || !cfg.Auth.Enabled) {
			return "", "", "", NewErrorWithStatusLookup(403, "API proxy requires header authentication")
		}
		remaining := rawPath[apiPrefixLen:]
		var user, repo string
		if strings.HasPrefix(remaining, "repos/") {
			parts := strings.SplitN(remaining[6:], "/", 3)
			if len(parts) >= 2 {
				user = parts[0]
				repo = parts[1]
			}
		} else if strings.HasPrefix(remaining, "users/") {
			parts := strings.SplitN(remaining[6:], "/", 2)
			if len(parts) >= 1 {
				user = parts[0]
			}
		}
		return user, repo, "api", nil
	}

	return "", "", "", NewErrorWithStatusLookup(404, "no matcher found for the given path")
}

var (
	proxyableMatchersMap map[string]struct{}
	initMatchersOnce     sync.Once
)

func initMatchers() {
	initMatchersOnce.Do(func() {
		matchers := []string{"blob", "raw", "gist"}
		proxyableMatchersMap = make(map[string]struct{}, len(matchers))
		for _, m := range matchers {
			proxyableMatchersMap[m] = struct{}{}
		}
	})
}

// matchString 与原始版本签名兼容
func matchString(target string) bool {
	initMatchers()
	_, exists := proxyableMatchersMap[target]
	return exists
}

// extractParts 与原始版本签名兼容
func extractParts(rawURL string) (string, string, string, url.Values, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", nil, err
	}

	path := parsedURL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	parts := strings.SplitN(path, "/", 3)

	if len(parts) < 2 {
		return "", "", "", nil, fmt.Errorf("URL path is too short")
	}

	repoOwner := "/" + parts[0]
	repoName := "/" + parts[1]
	var remainingPath string
	if len(parts) > 2 {
		remainingPath = "/" + parts[2]
	}

	return repoOwner, repoName, remainingPath, parsedURL.Query(), nil
}

var urlPattern = regexp.MustCompile(`https?://[^\s'"]+`)
