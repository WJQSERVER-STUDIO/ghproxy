package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthParametersHandler(c *app.RequestContext, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}

	authToken := c.Query("auth_token")
	logDebug("%s %s %s %s %s AUTH_TOKEN: %s", c.ClientIP(), c.Request.Method, string(c.Path()), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), authToken)

	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		return false, fmt.Errorf("Auth token invalid")
	}

	return isValid, nil
}
