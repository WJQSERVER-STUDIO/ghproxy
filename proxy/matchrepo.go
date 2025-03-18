package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
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

// processLinks 处理链接并返回一个 io.ReadCloser
func processLinks(input io.Reader, compress string, host string, cfg *config.Config) (io.ReadCloser, error) {
	var reader *bufio.Reader

	if compress == "gzip" {
		// 解压 gzip
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			return nil, fmt.Errorf("gzip 解压错误: %w", err)
		}
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(input)
	}

	// 创建一个缓冲区用于存储输出
	var outputBuffer io.Writer
	var gzipWriter *gzip.Writer
	var output io.ReadCloser
	var buf bytes.Buffer

	if compress == "gzip" {
		// 创建一个管道来连接 gzipWriter 和 output
		pipeReader, pipeWriter := io.Pipe() // 创建一个管道
		output = pipeReader                 // 将管道的读取端作为输出
		outputBuffer = pipeWriter           // 将管道的写入端作为 outputBuffer
		gzipWriter = gzip.NewWriter(outputBuffer)
		go func() {
			defer pipeWriter.Close() // 确保在 goroutine 结束时关闭 pipeWriter
			writer := bufio.NewWriter(gzipWriter)
			defer func() {
				if err := writer.Flush(); err != nil {
					logError("gzip writer 刷新失败: %v", err)
				}
				if err := gzipWriter.Close(); err != nil {
					logError("gzipWriter 关闭失败: %v", err)
				}
			}()

			scanner := bufio.NewScanner(reader)
			urlPattern := regexp.MustCompile(`https?://[^\s'"]+`)
			for scanner.Scan() {
				line := scanner.Text()
				modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
					return modifyURL(originalURL, host, cfg)
				})
				if _, err := writer.WriteString(modifiedLine + "\n"); err != nil {
					logError("写入 gzipWriter 失败: %v", err)
					return // 在发生错误时退出 goroutine
				}
			}
			if err := scanner.Err(); err != nil {
				logError("读取输入错误: %v", err)
			}
		}()
	} else {
		outputBuffer = &buf
		writer := bufio.NewWriter(outputBuffer)
		defer func() {
			if err := writer.Flush(); err != nil {
				logError("writer 刷新失败: %v", err)
			}
		}()

		urlPattern := regexp.MustCompile(`https?://[^\s'"]+`)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
				return modifyURL(originalURL, host, cfg)
			})
			if _, err := writer.WriteString(modifiedLine + "\n"); err != nil {
				return nil, fmt.Errorf("写入文件错误: %w", err) // 传递错误
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("读取行错误: %w", err) // 传递错误
		}
		output = io.NopCloser(&buf)
	}

	return output, nil
}

func WriteChunkedBody(resp io.ReadCloser, c *app.RequestContext) {
	defer resp.Close()

	c.Response.HijackWriter(hresp.NewChunkedBodyWriter(&c.Response, c.GetWriter()))

	bufWrapper := bytebufferpool.Get()
	buf := bufWrapper.B
	size := 32768 // 32KB
	buf = buf[:cap(buf)]
	if len(buf) < size {
		buf = append(buf, make([]byte, size-len(buf))...)
	}
	buf = buf[:size] // 将缓冲区限制为 'size'
	defer bytebufferpool.Put(bufWrapper)

	for {
		n, err := resp.Read(buf)
		if err != nil {
			if err == io.EOF {
				break // 读取到文件末尾
			}
			fmt.Println("读取错误:", err)
			c.String(http.StatusInternalServerError, "读取错误")
			return
		}

		_, err = c.Write(buf[:n]) // 写入 chunk
		if err != nil {
			fmt.Println("写入 chunk 错误:", err)
			return
		}

		c.Flush() // 刷新 chunk 到客户端
	}
}

// processLinksAndWriteChunked 处理链接并将结果以 chunked 方式写入响应
func ProcessLinksAndWriteChunked(input io.Reader, compress string, host string, cfg *config.Config, c *app.RequestContext) {
	var reader *bufio.Reader

	if compress == "gzip" {
		// 解压 gzip
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("gzip 解压错误: %v", err))
			return
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
			return // 发生错误时退出
		}

		if compress != "gzip" {
			if fErr := chunkedWriter.Flush(); fErr != nil {
				logError("chunkedWriter flush failed: %v", fErr)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logError("读取输入错误: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("读取输入错误: %v", err))
		return
	}

	// 对于 gzip，chunkedWriter 的关闭会触发最后的 chunk
	if compress != "gzip" {
		if fErr := chunkedWriter.Flush(); fErr != nil {
			logError("final chunkedWriter flush failed: %v", fErr)
		}
	}
}

func ProcessAndWriteChunkedBody(input io.Reader, compress string, host string, cfg *config.Config, c *app.RequestContext) error {
	var reader *bufio.Reader

	if compress == "gzip" {
		// 解压gzip
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			return fmt.Errorf("gzip解压错误: %v", err)
		}
		defer gzipReader.Close()
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(input)
	}

	// 创建一个缓冲区用于存储输出
	var outputBuffer io.Writer
	var gzipWriter *gzip.Writer
	var buf bytes.Buffer

	if compress == "gzip" {
		// 创建一个缓冲区
		outputBuffer = &buf
		gzipWriter = gzip.NewWriter(outputBuffer)
		defer func() {
			if gzipWriter != nil {
				if closeErr := gzipWriter.Close(); closeErr != nil {
					logError("gzipWriter close failed %v", closeErr)
				}
			}
		}()
	} else {
		outputBuffer = &buf
	}

	writer := bufio.NewWriter(outputBuffer)
	defer func() {
		if flushErr := writer.Flush(); flushErr != nil {
			logError("writer flush failed %v", flushErr)
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
			return fmt.Errorf("读取行错误: %v", err) // 传递错误
		}

		// 替换所有匹配的 URL
		modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
			return modifyURL(originalURL, host, cfg)
		})

		_, werr := writer.WriteString(modifiedLine)
		if werr != nil {
			return fmt.Errorf("写入文件错误: %v", werr) // 传递错误
		}
	}

	// 在返回之前，再刷新一次
	if fErr := writer.Flush(); fErr != nil {
		return fErr
	}

	if compress == "gzip" {
		if err := gzipWriter.Close(); err != nil {
			return fmt.Errorf("gzipWriter close failed: %v", err)
		}
	}

	// 将处理后的内容以分块的方式写入响应
	c.Response.HijackWriter(hresp.NewChunkedBodyWriter(&c.Response, c.GetWriter()))

	bufWrapper := bytebufferpool.Get()
	bbuf := bufWrapper.B
	size := 32768 // 32KB
	if cap(bbuf) < size {
		bbuf = make([]byte, size)
	} else {
		bbuf = bbuf[:size]
	}
	defer bytebufferpool.Put(bufWrapper)

	// 将缓冲区内容写入响应
	for {
		n, err := buf.Read(bbuf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("读取错误:", err)
				c.String(http.StatusInternalServerError, "读取错误")
				return err
			}
			break // 读取到文件末尾
		}

		_, err = c.Write(bbuf[:n]) // 写入 chunk
		if err != nil {
			fmt.Println("写入 chunk 错误:", err)
			return err
		}

		c.Flush() // 刷新 chunk 到客户端
	}

	return nil
}
