package auth

import (
	"fmt"
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
	if !cfg.Auth.Enabled {
		return true, ""
	}

	authToken := c.Query("auth_token")
	logInfo("%s %s %s %s %s AUTH_TOKEN: %s", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Request.Proto, authToken)

	if authToken == "" {
		err := "Auth token == nil"
		return false, err
	}

	isValid = authToken == cfg.Auth.AuthToken
	if !isValid {
		err := fmt.Sprintf("Auth token incorrect: %s", authToken)
		return false, err
	}

	logInfo("auth SUCCESS: %t", isValid)
	return isValid, ""
}
