package auth

import (
	"fmt"
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

// Auth Init
func Init(cfg *config.Config) {
	if cfg.Blacklist.Enabled {
		LoadBlacklist(cfg)
	}
	if cfg.Whitelist.Enabled {
		LoadWhitelist(cfg)
	}
	logInfo("Auth Init")
}

func AuthHandler(c *gin.Context, cfg *config.Config) (isValid bool, err string) {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth.Enabled {
		return true, ""
	}

	// 获取 auth_token 参数
	authToken := c.Query("auth_token")
	// IP METHOD URL USERAGENT PROTO TOKEN
	logInfo("%s %s %s %s %s AUTH_TOKEN: %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto, authToken)

	// 验证 token
	if authToken == "" {
		err := "Auth token == nil"
		return false, err
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		err := fmt.Sprintf("Auth token incorrect: %s", authToken)
		return false, err
	}

	logInfo("auth SUCCESS: %t", isValid)
	return isValid, ""
}
