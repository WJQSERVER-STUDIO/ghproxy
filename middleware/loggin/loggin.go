package loggin

import (
	"context"
	"time"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
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
		startTime := time.Now() // 请求开始处理前记录当前时间作为开始时间

		c.Next(ctx) //  调用 Next() 执行后续的 Handler

		endTime := time.Now()                   // 请求处理完成后记录当前时间作为结束时间
		timingResults := endTime.Sub(startTime) // 计算时间差，得到请求处理耗时 (Duration 类型)

		// 记录日志 IP METHOD URL USERAGENT PROTOCOL STATUS TIMING
		//  %s %s %s %s %s %d %s  分别对应:  ClientIP, Method, Protolcol, Path, UserAgent, StatusCode, timingResults (需要格式化)
		//  %v 可以通用地格式化 time.Duration 类型
		logInfo("%s %s %s %s %s %d %v ", c.ClientIP(), c.Method(), c.Request.Header.GetProtocol(), string(c.Path()), c.Request.Header.UserAgent(), c.Response.StatusCode(), timingResults)

		//logInfo("%s %s %s %s %d %v ", c.ClientIP(), c.Method(), c.Path(), c.Request.Header.UserAgent(), c.Response.StatusCode(), timingResults)
	}
}
