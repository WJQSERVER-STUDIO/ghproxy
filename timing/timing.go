package timing

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 阶段计时结构（固定数组优化）
type timingData struct {
	phases [8]struct { // 预分配8个阶段存储
		name string
		dur  time.Duration
	}
	count int
	start time.Time
}

// 对象池（内存重用优化）
var pool = sync.Pool{
	New: func() interface{} {
		return new(timingData)
	},
}

// 中间件入口
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从池中获取计时器
		td := pool.Get().(*timingData)
		td.start = time.Now()
		td.count = 0

		// 存储到上下文
		c.Set("timing", td)

		// 请求完成后回收对象
		defer func() {
			pool.Put(td)
		}()

		c.Next()
	}
}

// 记录阶段耗时
func Record(c *gin.Context, name string) {
	if val, exists := c.Get("timing"); exists {
		//td := val.(*timingData)
		td, ok := val.(*timingData)
		if !ok {
			return
		}
		if td.count < len(td.phases) {
			td.phases[td.count].name = name
			td.phases[td.count].dur = time.Since(td.start) // 直接记录当前时间
			td.count++
		}
	}
}

// 获取计时结果（日志输出用）
func Get(c *gin.Context) (total time.Duration, phases []struct {
	Name string
	Dur  time.Duration
}) {
	if val, exists := c.Get("timing"); exists {
		//td := val.(*timingData)
		td, ok := val.(*timingData)
		if !ok {
			return
		}
		for i := 0; i < td.count; i++ {
			phases = append(phases, struct {
				Name string
				Dur  time.Duration
			}{
				Name: td.phases[i].name,
				Dur:  td.phases[i].dur,
			})
		}
		total = time.Since(td.start)
	}
	return
}
