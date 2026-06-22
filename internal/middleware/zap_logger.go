package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgl "github.com/cyancyan2020/iam-platform/pkg/log"
	"go.uber.org/zap"
)

// TraceIDMiddleware 为每个请求注入 trace_id
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()[:8]
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}

// ZapLoggerMiddleware 替代 gin.Logger()，输出结构化请求日志
func ZapLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		traceID, _ := c.Get("trace_id")

		pkgl.Logger.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("trace_id", func() string {
				if id, ok := traceID.(string); ok {
					return id
				}
				return ""
			}()),
		)
	}
}
