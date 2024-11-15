package auth

import (
	"ghproxy/config"
	"ghproxy/logger"

	"github.com/gin-gonic/gin"
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func Init(cfg *config.Config) {
	if cfg.Blacklist.Enabled {
		LoadBlacklist(cfg)
	}
	if cfg.Whitelist.Enabled {
		LoadWhitelist(cfg)
	}
	logInfo("Auth Init")
}

func AuthHandler(c *gin.Context, cfg *config.Config) (isValid bool, err string) {
	if cfg.Auth.AuthMethod == "parameters" {
		isValid, err = AuthParametersHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.AuthMethod == "header" {
		isValid, err = AuthHeaderHandler(c, cfg)
		return isValid, err
	} else if cfg.Auth.AuthMethod == "" {
		logWarning("Auth method not set")
		return true, ""
	} else {
		logWarning("Auth method not supported")
		return false, "Auth method not supported"
	}
}
