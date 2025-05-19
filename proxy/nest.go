// Copyright 2025 WJQSERVER, WJQSERVER-STUDIO. All rights reserved.
// 使用本源代码受 WSL 2.0(WJQserver Studio License v2.0)与MPL 2.0(Mozilla Public License v2.0)许可协议的约束
// 此段代码使用双重授权许可, 允许用户选择其中一种许可证

package proxy

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"ghproxy/config"
	"io"
	"strings"
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
		logDump("Invalid URL: %s", url)
		return url
	}
	if matched {
		var u = url
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		logDump("Modified URL: %s", "https://"+host+"/"+u)
		return "https://" + host + "/" + u
	}
	return url
}

// processLinks 处理链接，返回包含处理后数据的 io.Reader
func processLinks(input io.ReadCloser, compress string, host string, cfg *config.Config) (readerOut io.Reader, written int64, err error) {
	pipeReader, pipeWriter := io.Pipe() // 创建 io.Pipe
	readerOut = pipeReader

	go func() { // 在 Goroutine 中执行写入操作
		defer func() {
			if pipeWriter != nil { // 确保 pipeWriter 关闭，即使发生错误
				if err != nil {
					if closeErr := pipeWriter.CloseWithError(err); closeErr != nil { // 如果有错误，传递错误给 reader
						logError("pipeWriter close with error failed: %v, original error: %v", closeErr, err)
					}
				} else {
					if closeErr := pipeWriter.Close(); closeErr != nil { // 没有错误，正常关闭
						logError("pipeWriter close failed: %v", closeErr)
						if err == nil { // 如果之前没有错误，记录关闭错误
							err = closeErr
						}
					}
				}
			}
		}()

		defer func() {
			if err := input.Close(); err != nil {
				logError("input close failed: %v", err)
			}

		}()

		var bufReader *bufio.Reader

		if compress == "gzip" {
			// 解压gzip
			gzipReader, gzipErr := gzip.NewReader(input)
			if gzipErr != nil {
				err = fmt.Errorf("gzip解压错误: %v", gzipErr)
				return // Goroutine 中使用 return 返回错误
			}
			defer gzipReader.Close()
			bufReader = bufio.NewReader(gzipReader)
		} else {
			bufReader = bufio.NewReader(input)
		}

		var bufWriter *bufio.Writer
		var gzipWriter *gzip.Writer

		// 根据是否gzip确定 writer 的创建
		if compress == "gzip" {
			gzipWriter = gzip.NewWriter(pipeWriter)           // 使用 pipeWriter
			bufWriter = bufio.NewWriterSize(gzipWriter, 4096) //设置缓冲区大小
		} else {
			bufWriter = bufio.NewWriterSize(pipeWriter, 4096) // 使用 pipeWriter
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
			if flushErr := bufWriter.Flush(); flushErr != nil {
				logError("writer flush failed %v", flushErr)
				// 如果已经存在错误，则保留。否则，记录此错误。
				if err == nil {
					err = flushErr
				}
			}
		}()

		// 使用正则表达式匹配 http 和 https 链接
		for {
			line, readErr := bufReader.ReadString('\n')
			if readErr != nil {
				if readErr == io.EOF {
					break // 文件结束
				}
				err = fmt.Errorf("读取行错误: %v", readErr) // 传递错误
				return                                 // Goroutine 中使用 return 返回错误
			}

			// 替换所有匹配的 URL
			modifiedLine := urlPattern.ReplaceAllStringFunc(line, func(originalURL string) string {
				logDump("originalURL: %s", originalURL)
				return modifyURL(originalURL, host, cfg) // 假设 modifyURL 函数已定义
			})

			n, writeErr := bufWriter.WriteString(modifiedLine)
			written += int64(n) // 更新写入的字节数
			if writeErr != nil {
				err = fmt.Errorf("写入文件错误: %v", writeErr) // 传递错误
				return                                   // Goroutine 中使用 return 返回错误
			}
		}

		// 在返回之前，再刷新一次 (虽然 defer 中已经有 flush，但这里再加一次确保及时刷新)
		if flushErr := bufWriter.Flush(); flushErr != nil {
			if err == nil { // 避免覆盖之前的错误
				err = flushErr
			}
			return // Goroutine 中使用 return 返回错误
		}
	}()

	return readerOut, written, nil // 返回 reader 和 written，error 由 Goroutine 通过 pipeWriter.CloseWithError 传递
}
