package repository

import (
	"context"
	"time"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"gorm.io/gorm"
)

type OperationLogRepository interface {
	Create(ctx context.Context, log *model.OperationLog) error
	Query(ctx context.Context, start, end time.Time, offset, limit int) ([]model.OperationLog, int64, error)
}

type operationLogRepository struct {
	db *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) OperationLogRepository {
	return &operationLogRepository{db: db}
}

func (r *operationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *operationLogRepository) Query(ctx context.Context, start, end time.Time, offset, limit int) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.OperationLog{})
	if !start.IsZero() {
		query = query.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("created_at <= ?", end)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
