package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/gin-gonic/gin"
)

func AuthParametersHandler(c *gin.Context, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}

	authToken := c.Query("auth_token")
	logDebug("%s %s %s %s %s AUTH_TOKEN: %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto, authToken)

	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		return false, fmt.Errorf("Auth token invalid")
	}

	return isValid, nil
}
