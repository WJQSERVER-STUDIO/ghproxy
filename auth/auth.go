package auth

import (
	"fmt"
	"ghproxy/config"

	"github.com/infinite-iroha/touka"
)

func Init(cfg *config.Config) {
	if cfg.Blacklist.Enabled {
		err := InitBlacklist(cfg)
		if err != nil {
			panic(err.Error())
		}
	}
	if cfg.Whitelist.Enabled {
		err := InitWhitelist(cfg)
		if err != nil {
			panic(err.Error())
		}
	}
}

func AuthHandler(c *touka.Context, cfg *config.Config) (isValid bool, err error) {
	if cfg.Auth.Method == "parameters" {
		isValid, err = AuthParametersHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.Method == "header" {
		isValid, err = AuthHeaderHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.Method == "" {
		c.Errorf("Auth method not set")
		return true, nil
	} else {
		c.Errorf("Auth method not supported %s", cfg.Auth.Method)
		return false, fmt.Errorf("%s", fmt.Sprintf("Auth method %s not supported", cfg.Auth.Method))
	}
}
