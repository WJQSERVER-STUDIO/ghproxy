package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

func AuthParametersHandler(c *touka.Context, cfg *config.Config) (isValid bool, err error) {
	if !cfg.Auth.Enabled {
		return true, nil
	}

	var authToken string
	if cfg.Auth.Key != "" {
		authToken = c.Query(cfg.Auth.Key)
	} else {
		authToken = c.Query("auth_token")
	}

	if authToken == "" {
		return false, fmt.Errorf("Auth token not found")
	}

	isValid = authToken == cfg.Auth.Token
	if !isValid {
		return false, fmt.Errorf("Auth token invalid")
	}

	return isValid, nil
}
