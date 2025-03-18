package auth

import (
	"context"
	"fmt"
	"ghproxy/config"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

var (
	logw       = logger.Logw
	LogDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func Init(cfg *config.Config) {
	if cfg.Blacklist.Enabled {
		err := InitBlacklist(cfg)
		if err != nil {
			logError(err.Error())
			return
		}
	}
	if cfg.Whitelist.Enabled {
		err := InitWhitelist(cfg)
		if err != nil {
			logError(err.Error())
			return
		}
	}
	logDebug("Auth Init")
}

func AuthHandler(ctx context.Context, c *app.RequestContext, cfg *config.Config) (isValid bool, err error) {
	if cfg.Auth.AuthMethod == "parameters" {
		isValid, err = AuthParametersHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.AuthMethod == "header" {
		isValid, err = AuthHeaderHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.AuthMethod == "" {
		logError("Auth method not set")
		return true, nil
	} else {
		logError("Auth method not supported")
		return false, fmt.Errorf(fmt.Sprintf("Auth method %s not supported", cfg.Auth.AuthMethod))
	}
}
