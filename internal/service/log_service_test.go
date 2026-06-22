package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
)

func TestLogQuery_Success(t *testing.T) {
	repo := mocks.NewOperationLogRepository(t)
	now := time.Now()
	repo.On("Query", mock.Anything, mock.Anything, mock.Anything, 0, 20).
		Return([]model.OperationLog{
			{ID: 1, Username: "admin", Method: "POST", Path: "/api/v1/users/login", StatusCode: 200, DurationMs: 45},
		}, int64(1), nil)

	svc := NewLogService(repo)
	result, err := svc.Query(context.Background(), &LogQuery{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("查询应成功: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("期望 1 条, 实际: %d", result.Total)
	}
	_ = now
}

func TestLogQuery_WithTimeRange(t *testing.T) {
	repo := mocks.NewOperationLogRepository(t)
	repo.On("Query", mock.Anything, mock.Anything, mock.Anything, 0, 20).
		Return([]model.OperationLog{}, int64(0), nil)

	svc := NewLogService(repo)
	result, err := svc.Query(context.Background(), &LogQuery{
		Page:  1,
		Size:  20,
		Start: "2026-01-01 00:00:00",
		End:   "2026-01-02 00:00:00",
	})
	if err != nil {
		t.Fatalf("带时间范围查询应成功: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("期望 0 条, 实际: %d", result.Total)
	}
}

func TestLogQuery_InvalidStartFormat(t *testing.T) {
	svc := NewLogService(mocks.NewOperationLogRepository(t))
	_, err := svc.Query(context.Background(), &LogQuery{
		Page:  1,
		Size:  20,
		Start: "2026/01/01",
	})
	if !errors.Is(err, ErrInvalidDateFormat) {
		t.Fatalf("期望 ErrInvalidDateFormat, 实际: %v", err)
	}
}

func TestLogQuery_InvalidEndFormat(t *testing.T) {
	svc := NewLogService(mocks.NewOperationLogRepository(t))
	_, err := svc.Query(context.Background(), &LogQuery{
		Page: 1,
		Size: 20,
		End:  "bad-date",
	})
	if !errors.Is(err, ErrInvalidDateFormat) {
		t.Fatalf("期望 ErrInvalidDateFormat, 实际: %v", err)
	}
}

func TestLogQuery_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := mocks.NewOperationLogRepository(t)
	repo.On("Query", mock.Anything, mock.Anything, mock.Anything, 0, 20).
		Return(nil, int64(0), dbErr)

	svc := NewLogService(repo)
	_, err := svc.Query(context.Background(), &LogQuery{Page: 1, Size: 20})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传 DB 错误, 实际: %v", err)
	}
}

func TestLogQuery_DefaultPaging(t *testing.T) {
	repo := mocks.NewOperationLogRepository(t)
	repo.On("Query", mock.Anything, mock.Anything, mock.Anything, 0, 20).
		Return([]model.OperationLog{}, int64(0), nil)

	svc := NewLogService(repo)
	_, err := svc.Query(context.Background(), &LogQuery{Page: 0, Size: 0})
	if err != nil {
		t.Fatalf("零值分页应使用默认值: %v", err)
	}
}
