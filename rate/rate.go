package rate

import (
	"time"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"golang.org/x/time/rate"
)

// 日志输出
var (
	logw       = logger.Logw
	LogDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

// 总体限流器
type RateLimiter struct {
	limiter *rate.Limiter
}

// 基于IP的限流器
type IPRateLimiter struct {
	limiters map[string]*RateLimiter
	limit    int
	burst    int
	duration time.Duration
}

func New(limit int, burst int, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(float64(limit)/duration.Seconds()), burst),
	}
}

func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

func NewIPRateLimiter(limit int, burst int, duration time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*RateLimiter),
		limit:    limit,
		burst:    burst,
		duration: duration,
	}
}

func (rl *IPRateLimiter) Allow(ip string) bool {
	if ip == "" {
		logWarning("empty ip")
		return false
	}

	limiter, ok := rl.limiters[ip]
	if !ok {
		// 创建新的 RateLimiter 并存储
		limiter = New(rl.limit, rl.burst, rl.duration)
		rl.limiters[ip] = limiter
	}
	return limiter.Allow()
}
