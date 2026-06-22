package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// incrWithExpire Lua 脚本：原子执行 INCR，首次时设置 TTL
var incrWithExpire = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])
	if count == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return count
`)

// RateLimiter 限流接口
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type redisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter 基于 Redis 的固定窗口限流器
// 注：当前为固定窗口，在窗口边界附近的极短时间内可通过最多 2×limit 次；
// 严苛场景可改用 ZSET 滑动窗口，当前实现满足登录保护需求。
func NewRedisRateLimiter(client *redis.Client) RateLimiter {
	return &redisRateLimiter{client: client}
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := incrWithExpire.Run(ctx, r.client, []string{key}, int(window.Seconds())).Int64()
	if err != nil {
		return false, err
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
