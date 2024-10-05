package auth

import (
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

var logw = logger.Logw

func AuthHandler(c *gin.Context, cfg *config.Config) bool {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth.Enabled {
		logw("auth PASSED")
		return true
	}

	// 获取 auth_token 参数
	authToken := c.Query("auth_token")
	logw("auth_token received: %s", authToken)

	// 验证 token
	if authToken == "" {
		logw("auth FAILED: no auth_token provided")
		return false
	}

	isValid := authToken == cfg.Auth.AuthToken
	if !isValid {
		logw("auth FAILED: invalid auth_token: %s", authToken)
	}

	return isValid
}

func IsBlacklisted(username, repo string, blacklist map[string][]string, enabled bool) bool {
	if !enabled {
		return false
	}

	// 检查 blacklist 是否为 nil
	if blacklist == nil {
		// 可以选择记录日志或返回 false
		logw("Warning: Blacklist map is nil")
		return false
	}

	if repos, ok := blacklist[username]; ok {
		for _, blacklistedRepo := range repos {
			if blacklistedRepo == repo {
				return true
			}
		}
	}

	return false
}
