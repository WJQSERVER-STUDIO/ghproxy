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
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/http1/resp"
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

func ProcessLinksAndWriteChunked(input io.Reader, compress string, host string, cfg *config.Config, c *app.RequestContext) error {
	pr, pw := io.Pipe() // 创建一个管道，用于进程间通信
	var wg sync.WaitGroup
	wg.Add(2)

	var processErr error // 用于存储处理过程中发生的错误

	go func() {
		defer wg.Done()  // 协程结束时通知 WaitGroup
		defer pw.Close() // 协程结束时关闭管道的写端

		var reader *bufio.Reader
		if compress == "gzip" { // 如果需要解压
			gzipReader, err := gzip.NewReader(input) // 创建 gzip 解压器
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("gzip 解压错误: %v", err)) // 设置 HTTP 状态码和错误信息
				processErr = fmt.Errorf("gzip decompression error: %w", err)                // gzip decompression error
				return
			}
			defer gzipReader.Close()             // 延迟关闭 gzip 解压器
			reader = bufio.NewReader(gzipReader) // 使用 bufio 读取解压后的数据
		} else {
			reader = bufio.NewReader(input) // 直接使用 bufio 读取原始数据
		}

		var writer io.Writer = pw // 默认写入管道
		var gzipWriter *gzip.Writer

		if compress == "gzip" { // 如果需要压缩
			gzipWriter = gzip.NewWriter(writer) // 创建 gzip 压缩器
			writer = gzipWriter                 // 将 writer 设置为 gzip 压缩器
			defer func() {                      // 延迟关闭 gzip 压缩器
				if err := gzipWriter.Close(); err != nil {
					logError("gzipWriter close failed: %v", err)
					processErr = fmt.Errorf("gzipwriter close failed: %w", err) // gzipwriter close failed
				}
			}()
		}

		urlPattern := regexp.MustCompile(`https?://[^\s'"]+`) // 编译正则表达式，用于匹配 URL
		scanner := bufio.NewScanner(reader)                   // 创建 scanner 用于逐行扫描
		for scanner.Scan() {                                  // 循环读取每一行
			line := scanner.Text()                                                                  // 获取当前行
			modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string { // 替换 URL
				return modifyURL(originalURL, host, cfg) // 调用 modifyURL 函数修改 URL
			})
			modifiedLineWithNewline := modifiedLine + "\n" // 添加换行符

			_, err := writer.Write([]byte(modifiedLineWithNewline)) // 将修改后的行写入管道/gzip
			if err != nil {
				logError("写入 pipe 错误: %v", err)                         // 记录错误
				processErr = fmt.Errorf("write to pipe error: %w", err) // write to pipe error
				return
			}
		}

		if err := scanner.Err(); err != nil {
			logError("读取输入错误: %v", err)                                              // 记录错误
			c.String(http.StatusInternalServerError, fmt.Sprintf("读取输入错误: %v", err)) // 设置 HTTP 状态码和错误信息
			processErr = fmt.Errorf("read input error: %w", err)                     // read input error
			return
		}
	}()

	go func() {
		defer wg.Done() // 协程结束时通知 WaitGroup

		c.Response.HijackWriter(resp.NewChunkedBodyWriter(&c.Response, c.GetWriter())) // 劫持 writer，启用分块编码

		bufWrapper := bytebufferpool.Get() // 从对象池获取 bytebuffer
		buf := bufWrapper.B
		size := 32768 // 32KB, 设置缓冲区大小
		buf = buf[:cap(buf)]
		if len(buf) < size {
			buf = append(buf, make([]byte, size-len(buf))...)
		}
		buf = buf[:size]                     // 将缓冲区限制为 'size'
		defer bytebufferpool.Put(bufWrapper) // 延迟将 bytebuffer 放回对象池

		for { // 循环读取和写入数据
			n, err := pr.Read(buf) // 从管道读取数据
			if err != nil {
				if err == io.EOF { // 如果读取到文件末尾
					if n > 0 { // 确保写入所有剩余数据
						_, err := c.Write(buf[:n]) // 写入最后的数据块
						if err != nil {
							processErr = fmt.Errorf("failed to write final chunk: %w", err) // failed to write final chunk
							break
						}
					}
					c.Flush() // 刷新缓冲区
					break     // 读取到文件末尾, 退出循环
				}
				logError("hwriter.Writer read error: %v", err) // 记录错误
				if processErr == nil {
					processErr = fmt.Errorf("failed to read from pipe: %w", err) // failed to read from pipe
					// 不要在这里设置 http status code. 如果 read 失败, process 协程可能还没有完成, 它可能正在尝试设置 status code. 两个地方都设置会导致 race condition.
				}
				break // 读取错误，退出循环
			}

			if n > 0 { // 只有在实际读取到数据时才写入
				_, err = c.Write(buf[:n]) // 将数据写入响应
				if err != nil {
					// 处理写入错误 (考虑记录日志并可能中止)
					logError("hwriter.Writer write error: %v", err)
					if processErr == nil { // 仅当 processErr 尚未设置时才设置.
						processErr = fmt.Errorf("failed to write chunk: %w", err) // failed to write chunk
					}
					break // 写入错误, 退出循环
				}

				// 在大多数情况下，考虑移除 Flush. 仅在 *真正* 需要时保留它。
				if err := c.Flush(); err != nil {
					// 更强大的 Flush() 错误处理
					c.AbortWithStatus(http.StatusInternalServerError) // 中止响应
					logError("hwriter.Writer flush error: %v", err)
					if processErr == nil {
						processErr = fmt.Errorf("failed to flush chunk: %w", err) // failed to flush chunk
					}
					break // 刷新错误, 退出循环
				}
			}
		}
	}()

	wg.Wait()         // 等待两个协程结束
	return processErr // 返回错误
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
