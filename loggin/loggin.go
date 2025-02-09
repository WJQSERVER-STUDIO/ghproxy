package loggin

import (
	"ghproxy/timing"
	"time"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-gonic/gin"
)

var (
	logw       = logger.Logw
	LogDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

// 日志中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		var timingResults time.Duration

		// 获取计时结果
		timingResults, _ = timing.Get(c)

		// 记录日志 IP METHOD URL USERAGENT PROTOCOL STATUS TIMING
		logInfo("%s %s %s %s %d %s ", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent(), c.Writer.Status(), timingResults)
	}
}
