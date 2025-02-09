package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/gin-gonic/gin"
)

func AuthParametersHandler(c *gin.Context, cfg *config.Config) (isValid bool, err string) {
	if !cfg.Auth.Enabled {
		return true, ""
	}

	authToken := c.Query("auth_token")
	logDebug("%s %s %s %s %s AUTH_TOKEN: %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto, authToken)

	if authToken == "" {
		err := "Auth token == nil"
		return false, err
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		err := fmt.Sprintf("Auth token incorrect: %s", authToken)
		return false, err
	}

	return isValid, ""
}
