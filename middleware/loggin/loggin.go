package loggin

import (
	"context"
	"time"

	"github.com/WJQSERVER-STUDIO/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

// 日志中间件
func Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		startTime := time.Now()

		c.Next(ctx)

		endTime := time.Now()
		timingResults := endTime.Sub(startTime)

		logInfo("%s %s %s %s %s %d %v ", c.ClientIP(), c.Method(), c.Request.Header.GetProtocol(), string(c.Path()), c.Request.Header.UserAgent(), c.Response.StatusCode(), timingResults)
	}
}
