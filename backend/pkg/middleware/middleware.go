package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

func CORS(allowedOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigins == "*" || origin == allowedOrigins || allowedOrigins == "" {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if allowedOrigins == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-CSRF-Token")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	last     time.Time
	tokens   float64
	capacity float64
	rate     float64
}

func newRateLimiter(rps int) *rateLimiter {
	if rps <= 0 {
		rps = 100
	}
	return &rateLimiter{capacity: float64(rps), tokens: float64(rps), rate: float64(rps)}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(rl.last).Seconds()
	rl.last = now
	rl.tokens += elapsed * rl.rate
	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}
	if rl.tokens < 1 {
		return false
	}
	rl.tokens--
	return true
}

var globalLimiter = newRateLimiter(100)

func RateLimit(rps int) gin.HandlerFunc {
	if rps > 0 {
		globalLimiter = newRateLimiter(rps)
	}
	return func(c *gin.Context) {
		if !globalLimiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func RequestLogger(logFn func(method, path string, status int, latency time.Duration)) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logFn(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}
