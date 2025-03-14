package proxy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-gonic/gin"
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

// 读取请求体
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logError("failed to read request body: %v", err)
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer c.Request.Body.Close()
	return body, nil
}

func HandleError(c *gin.Context, message string) {
	c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", message))
	logError(message)
}
