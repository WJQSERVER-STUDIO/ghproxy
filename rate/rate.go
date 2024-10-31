package rate

import (
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func New(limit int, burst int, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(float64(limit)/duration.Seconds()), burst),
	}
}

func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}
