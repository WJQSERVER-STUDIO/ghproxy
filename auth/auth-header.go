package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthHeaderHandler(c *app.RequestContext, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}
	// 获取"GH-Auth"的值
	authToken := string(c.GetHeader("GH-Auth"))
	logDebug("%s %s %s %s %s AUTH_TOKEN: %s", c.Request.Method, string(c.Path()), c.Request.Header.UserAgent(), c.Request.Header.GetProtocol(), authToken)
	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		return false, fmt.Errorf("Auth token incorrect")
	}

	return isValid, nil
}
