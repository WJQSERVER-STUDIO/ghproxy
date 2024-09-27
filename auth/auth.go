package auth

import (
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

var (
	cfg *config.Config
	log = logger.Logw
)

func AuthHandler(c *gin.Context) bool {
	// 如果身份验证未启用，直接返回 true
	if !cfg.Auth {
		log("auth PASS")
		return true
	}

	// 获取 auth_token 参数
	authToken := c.Query("auth_token")
	log("auth_token: ", authToken)

	// 验证 token
	isValid := authToken == cfg.AuthToken
	if !isValid {
		log("auth FAIL")
	}

	return isValid
}
