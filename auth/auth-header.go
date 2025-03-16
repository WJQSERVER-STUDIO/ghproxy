package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/gin-gonic/gin"
)

func AuthHeaderHandler(c *gin.Context, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}
	// 获取"GH-Auth"的值
	authToken := c.GetHeader("GH-Auth")
	logDebug("%s %s %s %s %s AUTH_TOKEN: %s", c.Request.Method, c.Request.Host, c.Request.URL.Path, c.Request.Proto, c.Request.RemoteAddr, authToken)
	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		return false, fmt.Errorf("Auth token incorrect")
	}

	return isValid, nil
}
