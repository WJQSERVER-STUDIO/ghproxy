package proxy

import (
	"fmt"
	"ghproxy/auth"
	"ghproxy/config"
	"io"

	"github.com/infinite-iroha/touka"
)

// CountingReader is a reader that counts the number of bytes read.
// CountingReader 是一个计算已读字节数的读取器.
type CountingReader struct {
	reader    io.Reader
	bytesRead int64
}

// NewCountingReader creates a new CountingReader.
// NewCountingReader 创建一个新的 CountingReader.
func NewCountingReader(reader io.Reader) *CountingReader {
	return &CountingReader{
		reader: reader,
	}
}

func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.reader.Read(p)
	cr.bytesRead += int64(n)
	return n, err
}

// BytesRead returns the number of bytes read.
// BytesRead 返回已读字节数.
func (cr *CountingReader) BytesRead() int64 {
	return cr.bytesRead
}

// Close closes the underlying reader if it implements io.Closer.
// 如果底层读取器实现了 io.Closer, 则关闭它.
func (cr *CountingReader) Close() error {
	if closer, ok := cr.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func listCheck(cfg *config.Config, c *touka.Context, user string, repo string, rawPath string) bool {
	if cfg.Auth.ForceAllowApi && cfg.Auth.ForceAllowApiPassList {
		return false
	}
	// 白名单检查
	if cfg.Whitelist.Enabled {
		whitelist := auth.CheckWhitelist(user, repo)
		if !whitelist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Whitelist Blocked repo: %s/%s", user, repo)))
			c.Infof("%s %s %s %s %s Whitelist Blocked repo: %s/%s", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, user, repo)
			return true
		}
	}

	// 黑名单检查
	if cfg.Blacklist.Enabled {
		blacklist := auth.CheckBlacklist(user, repo)
		if blacklist {
			ErrorPage(c, NewErrorWithStatusLookup(403, fmt.Sprintf("Blacklist Blocked repo: %s/%s", user, repo)))
			c.Infof("%s %s %s %s %s Blacklist Blocked repo: %s/%s", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, user, repo)
			return true
		}
	}

	return false
}

// 鉴权
func authCheck(c *touka.Context, cfg *config.Config, matcher string, rawPath string) bool {
	var err error

	if matcher == "api" && !cfg.Auth.ForceAllowApi {
		if cfg.Auth.Method != "header" || !cfg.Auth.Enabled {
			ErrorPage(c, NewErrorWithStatusLookup(403, "Github API Req without AuthHeader is Not Allowed"))
			c.Infof("%s %s %s AuthHeader Unavailable", c.ClientIP(), c.Request.Method, rawPath)
			return true
		}
	}

	// 鉴权
	if cfg.Auth.Enabled {
		var authcheck bool
		authcheck, err = auth.AuthHandler(c, cfg)
		if !authcheck {
			ErrorPage(c, NewErrorWithStatusLookup(401, fmt.Sprintf("Unauthorized: %v", err)))
			c.Infof("%s %s %s %s %s Auth-Error: %v", c.ClientIP(), c.Request.Method, rawPath, c.UserAgent(), c.Request.Proto, err)
			return true
		}
	}

	return false
}
