package auth

import (
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	LogWarning = logger.LogWarning
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

func AuthHandler(c *gin.Context, cfg *config.Config) bool {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth.Enabled {
		return true
	}

	// 获取 auth_token 参数
	authToken := c.Query("auth_token")
	logInfo("auth_token received: %s", authToken)

	// 验证 token
	if authToken == "" {
		LogWarning("auth FAILED: no auth_token provided")
		return false
	}

	isValid := authToken == cfg.Auth.AuthToken
	if !isValid {
		LogWarning("auth FAILED: invalid auth_token: %s", authToken)
	}

	logInfo("auth SUCCESS: %t", isValid)
	return isValid
}
