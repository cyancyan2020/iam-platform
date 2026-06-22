package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// mockRateLimiter 测试用限流器
type mockRateLimiter struct {
	allowed bool
	err     error
}

func (m *mockRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return m.allowed, m.err
}

// TestLoginRateLimit_Allowed 未超限正常放行
func TestLoginRateLimit_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", LoginRateLimitMiddleware(&mockRateLimiter{allowed: true}, 5, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// TestLoginRateLimit_Blocked 超限返回 429
func TestLoginRateLimit_Blocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", LoginRateLimitMiddleware(&mockRateLimiter{allowed: false}, 5, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("期望 429, 实际: %d", w.Code)
	}
	if w.Header().Get("Retry-After") != "60" {
		t.Fatalf("期望 Retry-After: 60, 实际: %s", w.Header().Get("Retry-After"))
	}
}

// TestLoginRateLimit_RedisError  Redis 异常时放行
func TestLoginRateLimit_RedisError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", LoginRateLimitMiddleware(&mockRateLimiter{err: context.DeadlineExceeded}, 5, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Redis 异常时应放行, 期望 200, 实际: %d", w.Code)
	}
}
