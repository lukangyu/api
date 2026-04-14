package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	r        rate.Limit
	b        int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	if r <= 0 {
		r = 20
	}
	if b <= 0 {
		b = 40
	}
	rl := &RateLimiter{visitors: map[string]*visitor{}, r: r, b: b}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetString(CtxApiKeyID)
		if key == "" {
			if v, ok := c.Get(CtxApiKeyID); ok {
				key = anyToString(v)
			}
		}
		if key == "" {
			c.Next()
			return
		}
		limiter := rl.getVisitor(key)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter) getVisitor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[key]
	if !ok {
		l := rate.NewLimiter(rl.r, rl.b)
		rl.visitors[key] = &visitor{limiter: l, lastSeen: time.Now()}
		return l
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func anyToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case uint:
		return uintToString(t)
	case int:
		return intToString(t)
	default:
		return ""
	}
}

func uintToString(v uint) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{byte('0' + (v % 10))}, buf...)
		v /= 10
	}
	return string(buf)
}

func intToString(v int) string {
	if v == 0 {
		return "0"
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{byte('0' + (v % 10))}, buf...)
		v /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
