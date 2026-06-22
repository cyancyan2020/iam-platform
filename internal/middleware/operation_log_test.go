package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/gin-gonic/gin"
)

func setupLogTestRouter(logChan chan model.OperationLog) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OperationLogMiddleware(logChan))
	r.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

// TestOperationLogMiddleware_CapturesRequest 正常请求被记录
func TestOperationLogMiddleware_CapturesRequest(t *testing.T) {
	logChan := make(chan model.OperationLog, 1)
	router := setupLogTestRouter(logChan)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}

	// channel 中应有 1 条日志
	select {
	case entry := <-logChan:
		if entry.Path != "/api/v1/test" {
			t.Fatalf("期望路径 /api/v1/test, 实际: %s", entry.Path)
		}
		if entry.Method != "GET" {
			t.Fatalf("期望方法 GET, 实际: %s", entry.Method)
		}
		if entry.StatusCode != http.StatusOK {
			t.Fatalf("期望状态码 200, 实际: %d", entry.StatusCode)
		}
	default:
		t.Fatal("channel 中应有日志记录")
	}
}

// TestOperationLogMiddleware_SkipsHealth 健康检查不被记录
func TestOperationLogMiddleware_SkipsHealth(t *testing.T) {
	logChan := make(chan model.OperationLog, 1)
	router := setupLogTestRouter(logChan)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}

	// channel 应为空（/health 被过滤）
	select {
	case entry := <-logChan:
		t.Fatalf("/health 不应被记录, 却记录了: %s %s", entry.Method, entry.Path)
	default:
		// 预期行为：无日志
	}
}

// TestOperationLogMiddleware_ChannelFullDoesNotBlock channel 满时不阻塞
func TestOperationLogMiddleware_ChannelFullDoesNotBlock(t *testing.T) {
	logChan := make(chan model.OperationLog) // 无缓冲
	// 先不启动 consumer，让 channel 满
	router := setupLogTestRouter(logChan)

	// 第一个请求写入 channel（无缓冲，阻塞直到被读取）
	// 用 goroutine 发送第一个请求，让它阻塞在 channel 写入
	done := make(chan bool, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		done <- true
	}()

	// 让第一个 goroutine 的 entry 进入 channel
	<-logChan

	// 现在 channel 已空，发第二个请求应从 channel 读取再阻塞
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		done <- true
	}()

	// 发第三个请求 — channel 应该满，default 丢弃，不阻塞
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("channel 满时请求不应被阻塞, 期望 200, 实际: %d", w3.Code)
	}

	// 清理
	<-logChan
	<-done
	<-done
}

// TestOperationLogMiddleware_StatusCode 验证状态码记录
func TestOperationLogMiddleware_StatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logChan := make(chan model.OperationLog, 1)
	r.Use(OperationLogMiddleware(logChan))
	r.GET("/api/v1/notfound", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notfound", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := <-logChan
	if entry.StatusCode != http.StatusNotFound {
		t.Fatalf("期望 404, 实际: %d", entry.StatusCode)
	}
}
