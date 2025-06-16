package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

func AuthHeaderHandler(c *touka.Context, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}
	// 获取"GH-Auth"的值
	var authToken string
	if cfg.Auth.Key != "" {
		authToken = string(c.Request.Header.Get(cfg.Auth.Key))

	} else {
		authToken = string(c.Request.Header.Get("GH-Auth"))
	}
	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.Token
	if !isValid {
		return false, fmt.Errorf("Auth token incorrect")
	}

	return isValid, nil
}
