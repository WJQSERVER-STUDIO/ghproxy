package proxy

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"ghproxy/config"
	"io"
	"regexp"
	"strings"
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

// processLinks 处理链接并将结果写入输出流
func processLinks(input io.Reader, output io.Writer, compress string, host string, cfg *config.Config) (written int64, err error) {
	var reader *bufio.Reader

	if compress == "gzip" {
		// 解压gzip
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			return 0, fmt.Errorf("gzip解压错误: %v", err)
		}
		defer gzipReader.Close()
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(input)
	}

	var writer *bufio.Writer
	var gzipWriter *gzip.Writer

	// 根据是否gzip确定 writer 的创建
	if compress == "gzip" {
		gzipWriter = gzip.NewWriter(output)
		writer = bufio.NewWriterSize(gzipWriter, 4096) //设置缓冲区大小
	} else {
		writer = bufio.NewWriterSize(output, 4096)
	}

	//确保writer关闭
	defer func() {
		var closeErr error // 局部变量，用于保存defer中可能发生的错误

		if gzipWriter != nil {
			if closeErr = gzipWriter.Close(); closeErr != nil {
				logError("gzipWriter close failed %v", closeErr)
				// 如果已经存在错误，则保留。否则，记录此错误。
				if err == nil {
					err = closeErr
				}
			}
		}
		if flushErr := writer.Flush(); flushErr != nil {
			logError("writer flush failed %v", flushErr)
			// 如果已经存在错误，则保留。否则，记录此错误。
			if err == nil {
				err = flushErr
			}
		}
	}()

	// 使用正则表达式匹配 http 和 https 链接
	urlPattern := regexp.MustCompile(`https?://[^\s'"]+`)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break // 文件结束
			}
			return written, fmt.Errorf("读取行错误: %v", err) // 传递错误
		}

		// 替换所有匹配的 URL
		modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
			return modifyURL(originalURL, host, cfg)
		})

		n, werr := writer.WriteString(modifiedLine)
		written += int64(n) // 更新写入的字节数
		if werr != nil {
			return written, fmt.Errorf("写入文件错误: %v", werr) // 传递错误
		}
	}

	// 在返回之前，再刷新一次
	if fErr := writer.Flush(); fErr != nil {
		return written, fErr
	}

	return written, nil
}
