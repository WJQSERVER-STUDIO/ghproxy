package proxy

import (
	"net/http"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

// 日志模块
var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func HandleError(c *app.RequestContext, message string) {
	c.JSON(http.StatusInternalServerError, map[string]string{"error": message})
	logError(message)
}
