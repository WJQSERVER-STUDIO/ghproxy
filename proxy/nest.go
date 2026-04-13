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
	prefixAPIBytes = []byte("https://api.github.com")
	prefixHTTP     = []byte("http://")
	prefixHTTPS    = []byte("https://")
)

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

func EditorMatcherBytes(rawPath []byte, cfg *config.Config) bool {
	if bytes.HasPrefix(rawPath, prefixGithub) {
		return true
	}
	if bytes.HasPrefix(rawPath, prefixRawUser) {
		return true
	}
	if bytes.HasPrefix(rawPath, prefixRaw) {
		return true
	}
	if bytes.HasPrefix(rawPath, prefixGistUser) {
		return true
	}
	if bytes.HasPrefix(rawPath, prefixGist) {
		return true
	}
	if cfg.Shell.RewriteAPI && bytes.HasPrefix(rawPath, prefixAPIBytes) {
		return true
	}
	return false
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

func modifyURLBytes(url []byte, host []byte, cfg *config.Config) []byte {
	if !EditorMatcherBytes(url, cfg) {
		return url
	}

	var trimmed []byte
	if bytes.HasPrefix(url, prefixHTTPS) {
		trimmed = url[len(prefixHTTPS):]
	} else if bytes.HasPrefix(url, prefixHTTP) {
		trimmed = url[len(prefixHTTP):]
	} else {
		trimmed = url
	}

	newURL := make([]byte, len(prefixHTTPS)+len(host)+1+len(trimmed))
	written := 0
	written += copy(newURL[written:], prefixHTTPS)
	written += copy(newURL[written:], host)
	written += copy(newURL[written:], []byte("/"))
	copy(newURL[written:], trimmed)

	return newURL
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func processLinksStreamingInternal(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context) (readerOut io.Reader, written int64, err error) {
	pipeReader, pipeWriter := io.Pipe()
	readerOut = pipeReader

	go func() {
		defer func() {
			if err != nil {
				_ = pipeWriter.CloseWithError(err)
				return
			}
			_ = pipeWriter.Close()
		}()
		defer func() {
			if closeErr := input.Close(); closeErr != nil && c != nil {
				c.Errorf("input close failed: %v", closeErr)
			}
		}()

		bufReader := bufio.NewReader(input)
		bufWriter := bufio.NewWriterSize(pipeWriter, 4096)
		defer func() {
			if flushErr := bufWriter.Flush(); flushErr != nil && err == nil {
				err = fmt.Errorf("flush writer failed: %w", flushErr)
			}
		}()

		for {
			line, readErr := bufReader.ReadString('\n')
			if readErr != nil && readErr != io.EOF {
				err = fmt.Errorf("read error: %w", readErr)
				return
			}

			if len(line) > 0 {
				modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
					return modifyURL(originalURL, host, cfg)
				})

				n, writeErr := bufWriter.WriteString(modifiedLine)
				written += int64(n)
				if writeErr != nil {
					err = fmt.Errorf("write error: %w", writeErr)
					return
				}
			}

			if readErr == io.EOF {
				break
			}
		}
	}()

	return readerOut, written, nil
}

func processLinks(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context, bodySize int) (readerOut io.Reader, written int64, err error) {
	const sizeThreshold = 256 * 1024

	if bodySize == -1 || bodySize > sizeThreshold {
		return processLinksStreamingInternal(input, host, cfg, c)
	}

	return processLinksBufferedInternal(input, host, cfg, c)
}

func processLinksBufferedInternal(input io.ReadCloser, host string, cfg *config.Config, c *touka.Context) (readerOut io.Reader, written int64, err error) {
	pipeReader, pipeWriter := io.Pipe()
	readerOut = pipeReader
	hostBytes := []byte(host)

	go func() {
		defer func() {
			if closeErr := input.Close(); closeErr != nil && c != nil {
				c.Errorf("input close failed: %v", closeErr)
			}
		}()
		defer func() {
			if err != nil {
				_ = pipeWriter.CloseWithError(err)
				return
			}
			_ = pipeWriter.Close()
		}()

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufferPool.Put(buf)

		if _, err = buf.ReadFrom(input); err != nil {
			err = fmt.Errorf("reading input failed: %w", err)
			return
		}

		modifiedBytes := urlPattern.ReplaceAllFunc(buf.Bytes(), func(originalURL []byte) []byte {
			return modifyURLBytes(originalURL, hostBytes, cfg)
		})

		var n int
		n, err = pipeWriter.Write(modifiedBytes)
		written = int64(n)
		if err != nil {
			err = fmt.Errorf("writing to pipe failed: %w", err)
		}
	}()

	return readerOut, written, nil
}
