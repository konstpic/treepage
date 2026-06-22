package middleware

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func redisAddrFromEnv() string {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		return v
	}
	return os.Getenv("REDIS_URL")
}

func newRedisLimiter(addr string, rps int) (*redisLimiter, error) {
	if rps <= 0 {
		rps = 100
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDBFromEnv(),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &redisLimiter{client: client, rps: rps}, nil
}

func redisDBFromEnv() int {
	n, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return n
}

type redisLimiter struct {
	client *redis.Client
	rps    int
}

func (rl *redisLimiter) allow(ctx context.Context, key string) bool {
	k := fmt.Sprintf("treepage:ratelimit:%s", key)
	pipe := rl.client.TxPipeline()
	incr := pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return true
	}
	return incr.Val() <= int64(rl.rps)
}

func RateLimitFromEnv(serviceName string, rps int) gin.HandlerFunc {
	if addr := redisAddrFromEnv(); addr != "" {
		if rl, err := newRedisLimiter(addr, rps); err == nil {
			return func(c *gin.Context) {
				key := c.ClientIP()
				if serviceName != "" {
					key = serviceName + ":" + key
				}
				if !rl.allow(c.Request.Context(), key) {
					c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
					return
				}
				c.Next()
			}
		}
	}
	return RateLimit(rps)
}
