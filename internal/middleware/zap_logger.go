package middleware

import (
	"strings"
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
			traceID = strings.ReplaceAll(uuid.New().String(), "-", "")[:12]
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}

// getTraceID 从 Context 安全提取 trace_id
func getTraceID(c *gin.Context) string {
	if id, ok := c.Get("trace_id"); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

// ZapLoggerMiddleware 替代 gin.Logger()，输出结构化请求日志
func ZapLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		pkgl.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("trace_id", getTraceID(c)),
		)
	}
}
