package nocache

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func NoCacheMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 设置禁止缓存的响应头
		c.Response.Header.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Response.Header.Set("Pragma", "no-cache")
		c.Response.Header.Set("Expires", "0")
		c.Next(ctx) // 继续处理请求
	}
}
