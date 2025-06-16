package nocache

import (
	"github.com/infinite-iroha/touka"
)

func NoCacheMiddleware() touka.HandlerFunc {
	return func(c *touka.Context) {
		// 设置禁止缓存的响应头
		c.SetHeader("Cache-Control", "no-store, no-cache, must-revalidate")
		c.SetHeader("Pragma", "no-cache")
		c.SetHeader("Expires", "0")
		c.Next() // 继续处理请求
	}
}
