package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
	mu       sync.Mutex
}

// RateLimiter implements a per-IP token bucket rate limiter.
type RateLimiter struct {
	visitors sync.Map
	rate     float64 // tokens refilled per second
	burst    float64 // maximum tokens (bucket size)
}

// NewRateLimiter creates a rate limiter. rps is requests per second, burst
// is the maximum burst size allowed.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rate:  rps,
		burst: float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	v, _ := rl.visitors.LoadOrStore(key, &visitor{tokens: rl.burst, lastSeen: time.Now()})
	vis := v.(*visitor)
	vis.mu.Lock()
	defer vis.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(vis.lastSeen).Seconds()
	vis.lastSeen = now
	vis.tokens += elapsed * rl.rate
	if vis.tokens > rl.burst {
		vis.tokens = rl.burst
	}

	if vis.tokens < 1 {
		return false
	}
	vis.tokens--
	return true
}

// RateLimit returns gin middleware that rejects requests exceeding the limit.
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			return
		}
		c.Next()
	}
}

// cleanup evicts visitors that haven't been seen in a while.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		threshold := time.Now().Add(-10 * time.Minute)
		rl.visitors.Range(func(key, value interface{}) bool {
			vis := value.(*visitor)
			vis.mu.Lock()
			if vis.lastSeen.Before(threshold) {
				rl.visitors.Delete(key)
			}
			vis.mu.Unlock()
			return true
		})
	}
}
