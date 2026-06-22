package middleware

import (
	"log"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件（异步写入）
func OperationLogMiddleware(logChan chan<- model.OperationLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// 提取用户信息（未登录也记录）
		userID := uint64(0)
		username := ""
		if val, ok := c.Get("user"); ok {
			if claims, ok := val.(*pkgjwt.Claims); ok {
				userID = claims.UserID
				username = claims.Username
			}
		}

		entry := model.OperationLog{
			UserID:     userID,
			Username:   username,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			DurationMs: int(time.Since(start).Milliseconds()),
			CreatedAt:  time.Now(),
		}

		// 异步写入，不阻塞请求响应
		select {
		case logChan <- entry:
		default:
			// channel 满时丢弃，避免阻塞请求
		}
	}
}

// LogConsumer 后台消费日志 channel，批量写入 MySQL
func LogConsumer(logRepo repository.OperationLogRepository, logChan <-chan model.OperationLog) {
	for entry := range logChan {
		if err := logRepo.Create(nil, &entry); err != nil {
			log.Printf("写入操作日志失败: %v", err)
		}
	}
}
