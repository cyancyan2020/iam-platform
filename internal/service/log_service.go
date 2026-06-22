package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
)

// ErrInvalidDateFormat 日期格式错误
var ErrInvalidDateFormat = errors.New("日期格式错误，示例: 2026-01-01 08:00:00")

type LogQuery struct {
	Start string `form:"start"` // "2006-01-02 15:04:05"
	End   string `form:"end"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

type LogListResult struct {
	List  []model.OperationLog `json:"list"`
	Total int64                `json:"total"`
}

type LogService interface {
	Query(ctx context.Context, query *LogQuery) (*LogListResult, error)
}

type logService struct {
	logRepo repository.OperationLogRepository
}

func NewLogService(logRepo repository.OperationLogRepository) LogService {
	return &logService{logRepo: logRepo}
}

func (s *logService) Query(ctx context.Context, query *LogQuery) (*LogListResult, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 || query.Size > 100 {
		query.Size = 20
	}
	offset := (query.Page - 1) * query.Size

	var start, end time.Time
	var err error
	if query.Start != "" {
		start, err = time.Parse("2006-01-02 15:04:05", query.Start)
		if err != nil {
			return nil, fmt.Errorf("%w: 开始日期 %s", ErrInvalidDateFormat, query.Start)
		}
	}
	if query.End != "" {
		end, err = time.Parse("2006-01-02 15:04:05", query.End)
		if err != nil {
			return nil, fmt.Errorf("%w: 结束日期 %s", ErrInvalidDateFormat, query.End)
		}
	}

	logs, total, err := s.logRepo.Query(ctx, start, end, offset, query.Size)
	if err != nil {
		return nil, err
	}
	return &LogListResult{List: logs, Total: total}, nil
}
