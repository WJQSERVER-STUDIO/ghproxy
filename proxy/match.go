package proxy

import (
	"fmt"
	"ghproxy/config"
	"net/url"
	"regexp"
	"strings"
)

func Matcher(rawPath string, cfg *config.Config) (string, string, string, *GHProxyErrors) {
	var (
		user    string
		repo    string
		matcher string
	)
	// 匹配 "https://github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://github.com") {
		remainingPath := strings.TrimPrefix(rawPath, "https://github.com")
		/*
			if strings.HasPrefix(remainingPath, "/") {
				remainingPath = strings.TrimPrefix(remainingPath, "/")
			}
		*/
		remainingPath = strings.TrimPrefix(remainingPath, "/")
		// 预期格式/user/repo/more...
		// 取出user和repo和最后部分
		parts := strings.Split(remainingPath, "/")
		if len(parts) <= 2 {
			errMsg := "Not enough parts in path after matching 'https://github.com*'"
			return "", "", "", NewErrorWithStatusLookup(400, errMsg)
		}
		user = parts[0]
		repo = parts[1]
		// 匹配 "https://github.com"开头的链接
		if len(parts) >= 3 {
			switch parts[2] {
			case "releases", "archive":
				matcher = "releases"
			case "blob":
				matcher = "blob"
			case "raw":
				matcher = "raw"
			case "info", "git-upload-pack":
				matcher = "clone"
			default:
				errMsg := "Url Matched 'https://github.com*', but didn't match the next matcher"
				return "", "", "", NewErrorWithStatusLookup(400, errMsg)
			}
		}
		return user, repo, matcher, nil
	}
	// 匹配 "https://raw"开头的链接
	if strings.HasPrefix(rawPath, "https://raw") {
		remainingPath := strings.TrimPrefix(rawPath, "https://")
		parts := strings.Split(remainingPath, "/")
		if len(parts) <= 3 {
			errMsg := "URL after matched 'https://raw*' should have at least 4 parts (user/repo/branch/file)."
			return "", "", "", NewErrorWithStatusLookup(400, errMsg)
		}
		user = parts[1]
		repo = parts[2]
		matcher = "raw"

		return user, repo, matcher, nil
	}
	// 匹配 "https://gist"开头的链接
	if strings.HasPrefix(rawPath, "https://gist") {
		remainingPath := strings.TrimPrefix(rawPath, "https://")
		parts := strings.Split(remainingPath, "/")
		if len(parts) <= 3 {
			errMsg := "URL after matched 'https://gist*' should have at least 4 parts (user/gist_id)."
			return "", "", "", NewErrorWithStatusLookup(400, errMsg)
		}
		user = parts[1]
		repo = ""
		matcher = "gist"
		return user, repo, matcher, nil
	}
	// 匹配 "https://api.github.com/"开头的链接
	if strings.HasPrefix(rawPath, "https://api.github.com/") {
		matcher = "api"
		remainingPath := strings.TrimPrefix(rawPath, "https://api.github.com/")

		parts := strings.Split(remainingPath, "/")
		if parts[0] == "repos" {
			user = parts[1]
			repo = parts[2]
		}
		if parts[0] == "users" {
			user = parts[1]
		}
		if !cfg.Auth.ForceAllowApi {
			if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
				//return "", "", "", ErrAuthHeaderUnavailable
				errMsg := "AuthHeader Unavailable, Need to open header auth to enable api proxy"
				return "", "", "", NewErrorWithStatusLookup(403, errMsg)
			}
		}
		return user, repo, matcher, nil
	}
	//return "", "", "", ErrNotFound
	errMsg := "Didn't match any matcher"
	return "", "", "", NewErrorWithStatusLookup(404, errMsg)
}

var (
	matchedMatchers = []string{
		"blob",
		"raw",
		"gist",
	}
)

// matchString 检查目标字符串是否在给定的字符串集合中
func matchString(target string, stringsToMatch []string) bool {
	matchMap := make(map[string]struct{}, len(stringsToMatch))
	for _, str := range stringsToMatch {
		matchMap[str] = struct{}{}
	}
	_, exists := matchMap[target]
	return exists
}

// extractParts 从给定的 URL 中提取所需的部分
func extractParts(rawURL string) (string, string, string, url.Values, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", nil, err
	}

	// 获取路径部分并分割
	pathParts := strings.Split(parsedURL.Path, "/")

	// 提取所需的部分
	if len(pathParts) < 3 {
		return "", "", "", nil, fmt.Errorf("URL path is too short")
	}

	// 提取 /WJQSERVER-STUDIO 和 /go-utils.git
	repoOwner := "/" + pathParts[1]
	repoName := "/" + pathParts[2]

	// 剩余部分
	remainingPath := strings.Join(pathParts[3:], "/")
	if remainingPath != "" {
		remainingPath = "/" + remainingPath
	}

	// 查询参数
	queryParams := parsedURL.Query()

	return repoOwner, repoName, remainingPath, queryParams, nil
}

var urlPattern = regexp.MustCompile(`https?://[^\s'"]+`)
