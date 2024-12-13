package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/gin-gonic/gin"
)

func AuthHeaderHandler(c *gin.Context, cfg *config.Config) (isValid bool, err string) {
	if !cfg.Auth.Enabled {
		return true, ""
	}
	// 获取"GH-Auth"的值
	authToken := c.GetHeader("GH-Auth")
	logInfo("%s %s %s %s %s AUTH_TOKEN: %s", c.Request.Method, c.Request.Host, c.Request.URL.Path, c.Request.Proto, c.Request.RemoteAddr, authToken)
	if authToken == "" {
		err := "Auth Header == nil"
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
