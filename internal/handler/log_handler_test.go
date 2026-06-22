package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/service"
	pkgl "github.com/cyancyan2020/iam-platform/pkg/log"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	pkgl.Init("debug")
	code := m.Run()
	os.Exit(code)
}

// ——— mock LogService ———

type mockLogService struct {
	queryFn func(ctx context.Context, query *service.LogQuery) (*service.LogListResult, error)
}

func (m *mockLogService) Query(ctx context.Context, query *service.LogQuery) (*service.LogListResult, error) {
	return m.queryFn(ctx, query)
}

func setupLogHandlerTestRouter(svc service.LogService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewLogHandler(svc)
	r.GET("/api/v1/logs", h.Query)
	return r
}

// ——— 200 正常查询 ———

func TestLogHandler_QuerySuccess(t *testing.T) {
	svc := &mockLogService{
		queryFn: func(ctx context.Context, query *service.LogQuery) (*service.LogListResult, error) {
			return &service.LogListResult{List: nil, Total: 0}, nil
		},
	}
	router := setupLogHandlerTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?page=1&size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// ——— 400 日期格式错误 ———

func TestLogHandler_QueryInvalidDate(t *testing.T) {
	svc := &mockLogService{
		queryFn: func(ctx context.Context, query *service.LogQuery) (*service.LogListResult, error) {
			return nil, service.ErrInvalidDateFormat
		},
	}
	router := setupLogHandlerTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?start=bad", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// ——— 500 DB 错误 ———

func TestLogHandler_QueryDBError(t *testing.T) {
	svc := &mockLogService{
		queryFn: func(ctx context.Context, query *service.LogQuery) (*service.LogListResult, error) {
			return nil, errors.New("connection refused")
		},
	}
	router := setupLogHandlerTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?page=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("期望 500, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	// 确认不泄露内部错误信息
	if strings.Contains(w.Body.String(), "connection refused") {
		t.Fatal("500 响应不应泄露内部错误信息")
	}
}
