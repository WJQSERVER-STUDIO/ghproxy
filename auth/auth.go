package auth

import (
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

var logw = logger.Logw

func AuthHandler(c *gin.Context, cfg *config.Config) bool {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth {
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

	isValid := authToken == cfg.AuthToken
	if !isValid {
		logw("auth FAILED: invalid auth_token: %s", authToken)
	}

	return isValid
}
