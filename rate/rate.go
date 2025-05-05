package rate

import (
	"sync"
	"time"

	"github.com/WJQSERVER-STUDIO/logger"
	"golang.org/x/time/rate"
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

// RateLimiter 总体限流器
type RateLimiter struct {
	limiter *rate.Limiter
}

// New 创建一个总体限流器
func New(limit int, burst int, duration time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 1
		logWarning("rate limit per minute must be positive, setting to 1")
	}
	if burst <= 0 {
		burst = 1
		logWarning("rate limit burst must be positive, setting to 1")
	}

	rateLimit := rate.Limit(float64(limit) / duration.Seconds())

	return &RateLimiter{
		limiter: rate.NewLimiter(rateLimit, burst),
	}
}

// Allow 检查是否允许请求通过
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// IPRateLimiter 基于IP的限流器
type IPRateLimiter struct {
	limiters map[string]*RateLimiter // 用户级限流器 map
	mu       sync.RWMutex            // 保护 limiters map
	limit    int                     // 每 duration 时间段内允许的请求数
	burst    int                     // 突发请求数
	duration time.Duration           // 限流周期
}

// NewIPRateLimiter 创建一个基于IP的限流器
func NewIPRateLimiter(ipLimit int, ipBurst int, duration time.Duration) *IPRateLimiter {
	if ipLimit <= 0 {
		ipLimit = 1
		logWarning("IP rate limit per minute must be positive, setting to 1")
	}
	if ipBurst <= 0 {
		ipBurst = 1
		logWarning("IP rate limit burst must be positive, setting to 1")
	}

	logInfo("IP Rate Limiter initialized with limit: %d, burst: %d, duration: %v", ipLimit, ipBurst, duration)

	return &IPRateLimiter{
		limiters: make(map[string]*RateLimiter),
		limit:    ipLimit,
		burst:    ipBurst,
		duration: duration,
	}
}

// Allow 检查给定IP的请求是否允许通过
func (rl *IPRateLimiter) Allow(ip string) bool {
	if ip == "" {
		logWarning("empty ip for rate limiting")
		return false
	}

	// 使用读锁快速查找
	rl.mu.RLock()
	limiter, found := rl.limiters[ip]
	rl.mu.RUnlock()

	if found {
		return limiter.Allow()
	}

	// 未找到，获取写锁来创建和添加
	rl.mu.Lock()
	// 双重检查
	limiter, found = rl.limiters[ip]
	if !found {
		newL := New(rl.limit, rl.burst, rl.duration)
		rl.limiters[ip] = newL
		limiter = newL
	}
	rl.mu.Unlock()

	return limiter.Allow()
}
