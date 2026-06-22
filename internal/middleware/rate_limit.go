package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流接口
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type redisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter 基于 Redis 的滑动窗口限流器
func NewRedisRateLimiter(client *redis.Client) RateLimiter {
	return &redisRateLimiter{client: client}
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	// 首次请求设置过期时间
	if count == 1 {
		r.client.Expire(ctx, key, window)
	}
	return count <= int64(limit), nil
}

// LoginRateLimitMiddleware 登录接口限流中间件
func LoginRateLimitMiddleware(limiter RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "login_rate:" + c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			// Redis 异常时放行，避免误阻断
			c.Next()
			return
		}
		if !allowed {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后重试",
			})
			return
		}
		c.Next()
	}
}
