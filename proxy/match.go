package proxy

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	hresp "github.com/cloudwego/hertz/pkg/protocol/http1/resp"
	"github.com/valyala/bytebufferpool"
)

// 定义错误类型, error承载描述, 便于处理
type MatcherErrors struct {
	Code int
	Msg  string
	Err  error
}

var (
	ErrInvalidURL = &MatcherErrors{
		Code: 403,
		Msg:  "Invalid URL Format",
	}
	ErrAuthHeaderUnavailable = &MatcherErrors{
		Code: 403,
		Msg:  "AuthHeader Unavailable",
	}
)

func (e *MatcherErrors) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Code: %d, Msg: %s, Err: %s", e.Code, e.Msg, e.Err.Error())
	}
	return fmt.Sprintf("Code: %d, Msg: %s", e.Code, e.Msg)
}

func (e *MatcherErrors) Unwrap() error {
	return e.Err
}

func Matcher(rawPath string, cfg *config.Config) (string, string, string, error) {
	var (
		user    string
		repo    string
		matcher string
	)
	// 匹配 "https://github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://github.com") {
		remainingPath := strings.TrimPrefix(rawPath, "https://github.com")
		if strings.HasPrefix(remainingPath, "/") {
			remainingPath = strings.TrimPrefix(remainingPath, "/")
		}
		// 预期格式/user/repo/more...
		// 取出user和repo和最后部分
		parts := strings.Split(remainingPath, "/")
		if len(parts) <= 2 {
			return "", "", "", ErrInvalidURL
		}
		user = parts[0]
		repo = parts[1]
		// 匹配 "https://github.com"开头的链接
		if len(parts) >= 3 {
			switch parts[2] {
			case "releases", "archive":
				matcher = "releases"
			case "blob", "raw":
				matcher = "blob"
			case "info", "git-upload-pack":
				matcher = "clone"
			default:
				return "", "", "", ErrInvalidURL
			}
		}
		return user, repo, matcher, nil
	}
	// 匹配 "https://raw"开头的链接
	if strings.HasPrefix(rawPath, "https://raw") {
		remainingPath := strings.TrimPrefix(rawPath, "https://")
		parts := strings.Split(remainingPath, "/")
		if len(parts) <= 3 {
			return "", "", "", ErrInvalidURL
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
			return "", "", "", ErrInvalidURL
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
			if cfg.Auth.AuthMethod != "header" || !cfg.Auth.Enabled {
				return "", "", "", ErrAuthHeaderUnavailable
			}
		}
		return user, repo, matcher, nil
	}
	return "", "", "", ErrInvalidURL
}

func EditorMatcher(rawPath string, cfg *config.Config) (bool, string, error) {
	var (
		matcher string
	)
	// 匹配 "https://github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://github.com") {
		remainingPath := strings.TrimPrefix(rawPath, "https://github.com")
		if strings.HasPrefix(remainingPath, "/") {
			remainingPath = strings.TrimPrefix(remainingPath, "/")
		}
		return true, "", nil
	}
	// 匹配 "https://raw.githubusercontent.com"开头的链接
	if strings.HasPrefix(rawPath, "https://raw.githubusercontent.com") {
		return true, matcher, nil
	}
	// 匹配 "https://raw.github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://raw.github.com") {
		return true, matcher, nil
	}
	// 匹配 "https://gist.githubusercontent.com"开头的链接
	if strings.HasPrefix(rawPath, "https://gist.githubusercontent.com") {
		return true, matcher, nil
	}
	// 匹配 "https://gist.github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://gist.github.com") {
		return true, matcher, nil
	}
	// 匹配 "https://api.github.com/"开头的链接
	if strings.HasPrefix(rawPath, "https://api.github.com") {
		matcher = "api"
		return true, matcher, nil
	}
	return false, "", ErrInvalidURL
}

// 匹配文件扩展名是sh的rawPath
func MatcherShell(rawPath string) bool {
	if strings.HasSuffix(rawPath, ".sh") {
		return true
	}
	return false
}

// LinkProcessor 是一个函数类型，用于处理提取到的链接。
type LinkProcessor func(string) string

// 自定义 URL 修改函数
func modifyURL(url string, host string, cfg *config.Config) string {
	// 去除url内的https://或http://
	matched, _, err := EditorMatcher(url, cfg)
	if err != nil {
		logDump("Invalid URL: %s", url)
		return url
	}
	if matched {

		u := strings.TrimPrefix(url, "https://")
		u = strings.TrimPrefix(url, "http://")
		logDump("Modified URL: %s", "https://"+host+"/"+u)
		return "https://" + host + "/" + u
	}
	return url
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

// processLinksAndWriteChunked 处理链接并将结果以 chunked 方式写入响应
func ProcessLinksAndWriteChunked(input io.Reader, compress string, host string, cfg *config.Config, c *app.RequestContext) error {
	var reader *bufio.Reader

	if compress == "gzip" {
		// 解压 gzip
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("gzip 解压错误: %v", err))
			return fmt.Errorf("gzip 解压错误: %w", err)
		}
		defer gzipReader.Close()
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(input)
	}

	// 获取 chunked body writer
	chunkedWriter := hresp.NewChunkedBodyWriter(&c.Response, c.GetWriter())

	var writer io.Writer = chunkedWriter
	var gzipWriter *gzip.Writer

	if compress == "gzip" {
		gzipWriter = gzip.NewWriter(writer)
		writer = gzipWriter
		defer func() {
			if err := gzipWriter.Close(); err != nil {
				logError("gzipWriter close failed: %v", err)
			}
		}()
	}

	bufWrapper := bytebufferpool.Get()
	buf := bufWrapper.B
	size := 32768 // 32KB
	buf = buf[:cap(buf)]
	if len(buf) < size {
		buf = append(buf, make([]byte, size-len(buf))...)
	}
	buf = buf[:size] // 将缓冲区限制为 'size'
	defer bytebufferpool.Put(bufWrapper)

	urlPattern := regexp.MustCompile(`https?://[^\s'"]+`)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
			return modifyURL(originalURL, host, cfg)
		})
		modifiedLineWithNewline := modifiedLine + "\n"

		_, err := writer.Write([]byte(modifiedLineWithNewline))
		if err != nil {
			logError("写入 chunk 错误: %v", err)
			return fmt.Errorf("写入 chunk 错误: %w", err)
		}

		if compress != "gzip" {
			if fErr := chunkedWriter.Flush(); fErr != nil {
				logError("chunkedWriter flush failed: %v", fErr)
				return fmt.Errorf("chunkedWriter flush failed: %w", fErr)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logError("读取输入错误: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("读取输入错误: %v", err))
		return fmt.Errorf("读取输入错误: %w", err)
	}

	// 对于 gzip，chunkedWriter 的关闭会触发最后的 chunk
	if compress != "gzip" {
		if fErr := chunkedWriter.Flush(); fErr != nil {
			logError("final chunkedWriter flush failed: %v", fErr)
			return fmt.Errorf("final chunkedWriter flush failed: %w", fErr)
		}
	}

	return nil // 成功完成处理
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
