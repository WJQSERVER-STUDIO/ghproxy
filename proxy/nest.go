package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"ghproxy/config"
	"io"
	"strings"
	"sync"

	"github.com/infinite-iroha/touka"
)

var (
	prefixGithub   = []byte("https://github.com")
	prefixRawUser  = []byte("https://raw.githubusercontent.com")
	prefixRaw      = []byte("https://raw.github.com")
	prefixGistUser = []byte("https://gist.githubusercontent.com")
	prefixGist     = []byte("https://gist.github.com")
	prefixAPI      = []byte("https://api.github.com")
	prefixHTTP     = []byte("http://")
	prefixHTTPS    = []byte("https://")
)

func EditorMatcherBytes(rawPath []byte, cfg *config.Config) (bool, error) {
	if bytes.HasPrefix(rawPath, prefixGithub) {
		return true, nil
	}
	if bytes.HasPrefix(rawPath, prefixRawUser) {
		return true, nil
	}
	if bytes.HasPrefix(rawPath, prefixRaw) {
		return true, nil
	}
	if bytes.HasPrefix(rawPath, prefixGistUser) {
		return true, nil
	}
	if bytes.HasPrefix(rawPath, prefixGist) {
		return true, nil
	}
	if cfg.Shell.RewriteAPI {
		if bytes.HasPrefix(rawPath, prefixAPI) {
			return true, nil
		}
	}
	return false, nil
}

func modifyURLBytes(url []byte, host []byte, cfg *config.Config) []byte {
	matched, err := EditorMatcherBytes(url, cfg)
	if err != nil || !matched {
		return url
	}

	var u []byte
	if bytes.HasPrefix(url, prefixHTTPS) {
		u = url[len(prefixHTTPS):]
	} else if bytes.HasPrefix(url, prefixHTTP) {
		u = url[len(prefixHTTP):]
	} else {
		u = url
	}

	newLen := len(prefixHTTPS) + len(host) + 1 + len(u)
	newURL := make([]byte, newLen)

	written := 0
	written += copy(newURL[written:], prefixHTTPS)
	written += copy(newURL[written:], host)
	written += copy(newURL[written:], []byte("/"))
	copy(newURL[written:], u)

	return newURL
}

func EditorMatcher(rawPath string, cfg *config.Config) (bool, error) {
	// 匹配 "https://github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://github.com") {
		return true, nil
	}
	// 匹配 "https://raw.githubusercontent.com"开头的链接
	if strings.HasPrefix(rawPath, "https://raw.githubusercontent.com") {
		return true, nil
	}
	// 匹配 "https://raw.github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://raw.github.com") {
		return true, nil
	}
	// 匹配 "https://gist.githubusercontent.com"开头的链接
	if strings.HasPrefix(rawPath, "https://gist.githubusercontent.com") {
		return true, nil
	}
	// 匹配 "https://gist.github.com"开头的链接
	if strings.HasPrefix(rawPath, "https://gist.github.com") {
		return true, nil
	}
	if cfg.Shell.RewriteAPI {
		// 匹配 "https://api.github.com/"开头的链接
		if strings.HasPrefix(rawPath, "https://api.github.com") {
			return true, nil
		}
	}
	return false, nil
}

// 匹配文件扩展名是sh的rawPath
func MatcherShell(rawPath string) bool {
	return strings.HasSuffix(rawPath, ".sh")
}

// LinkProcessor 是一个函数类型，用于处理提取到的链接。
type LinkProcessor func(string) string

// 自定义 URL 修改函数
func modifyURL(url string, host string, cfg *config.Config) string {
	// 去除url内的https://或http://
	matched, err := EditorMatcher(url, cfg)
	if err != nil {
		return url
	}
	if matched {
		var u = url
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		return "https://" + host + "/" + u
	}
	return url
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// processLinksStreamingInternal is a link processing function that reads the input line by line.
// It is memory-safe for large files but less performant due to numerous small allocations.
func processLinksStreamingInternal(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context) (readerOut io.Reader, written int64, err error) {
	pipeReader, pipeWriter := io.Pipe()
	readerOut = pipeReader

	go func() {
		defer func() {
			if err != nil {
				pipeWriter.CloseWithError(err)
			} else {
				pipeWriter.Close()
			}
		}()
		defer input.Close()

		bufReader := bufio.NewReader(input)
		bufWriter := bufio.NewWriterSize(pipeWriter, 4096)
		defer bufWriter.Flush()

		for {
			line, readErr := bufReader.ReadString('\n')
			if readErr != nil && readErr != io.EOF {
				err = fmt.Errorf("read error: %w", readErr)
				return
			}

			modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
				return modifyURL(originalURL, host, cfg)
			})

			var n int
			n, err = bufWriter.WriteString(modifiedLine)
			written += int64(n)
			if err != nil {
				err = fmt.Errorf("write error: %w", err)
				return
			}

			if readErr == io.EOF {
				break
			}
		}
	}()

	return readerOut, written, nil
}

// processLinks acts as a dispatcher, choosing the best processing strategy based on file size.
// It uses a memory-safe streaming approach for large or unknown-size files,
// and a high-performance buffered approach for smaller files.
func processLinks(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context, bodySize int) (readerOut io.Reader, written int64, err error) {
	const sizeThreshold = 256 * 1024 // 256KB

	// Use streaming for large or unknown size files to prevent OOM
	if bodySize == -1 || bodySize > sizeThreshold {
		c.Debugf("Using streaming processor for large/unknown size file (%d bytes)", bodySize)
		return processLinksStreamingInternal(input, host, cfg, c)
	} else {
		c.Debugf("Using buffered processor for small file (%d bytes)", bodySize)
		return processLinksBufferedInternal(input, host, cfg, c)
	}
}

// processLinksBufferedInternal a link processing function that reads the entire content into a buffer.
// It is optimized for performance on smaller files but carries an OOM risk for large files.
func processLinksBufferedInternal(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context) (readerOut io.Reader, written int64, err error) {
	pipeReader, pipeWriter := io.Pipe()
	readerOut = pipeReader
	hostBytes := []byte(host)

	go func() {
		// 在 goroutine 退出时, 根据 err 是否为 nil, 带错误或正常关闭 pipeWriter
		defer func() {
			if closeErr := input.Close(); closeErr != nil {
				c.Errorf("input close failed: %v", closeErr)
			}
		}()
		defer func() {
			if err != nil {
				if closeErr := pipeWriter.CloseWithError(err); closeErr != nil {
					c.Errorf("pipeWriter close with error failed: %v", closeErr)
				}
			} else {
				if closeErr := pipeWriter.Close(); closeErr != nil {
					c.Errorf("pipeWriter close failed: %v", closeErr)
				}
			}
		}()

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufferPool.Put(buf)

		// 将全部输入读入复用的缓冲区
		if _, err = buf.ReadFrom(input); err != nil {
			err = fmt.Errorf("reading input failed: %w", err)
			return
		}

		// 使用 ReplaceAllFunc 和字节版本辅助函数, 实现准零分配
		modifiedBytes := urlPattern.ReplaceAllFunc(buf.Bytes(), func(originalURL []byte) []byte {
			return modifyURLBytes(originalURL, hostBytes, cfg)
		})

		// 将处理后的字节写回管道
		var n int
		n, err = pipeWriter.Write(modifiedBytes)
		if err != nil {
			err = fmt.Errorf("writing to pipe failed: %w", err)
			return
		}
		written = int64(n)
	}()

	return readerOut, written, nil
}
