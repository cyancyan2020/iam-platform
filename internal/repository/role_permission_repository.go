package repository

import (
	"context"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"gorm.io/gorm"
)

type RolePermissionRepository interface {
	FindByRoleID(ctx context.Context, roleID uint64) ([]model.RolePermission, error)
	BatchReplace(ctx context.Context, roleID uint64, permIDs []uint64) error
}

type rolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) FindByRoleID(ctx context.Context, roleID uint64) ([]model.RolePermission, error) {
	var rps []model.RolePermission
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rps).Error
	return rps, err
}

func (r *rolePermissionRepository) BatchReplace(ctx context.Context, roleID uint64, permIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permIDs) == 0 {
			return nil
		}
		rps := make([]model.RolePermission, 0, len(permIDs))
		for _, pid := range permIDs {
			rps = append(rps, model.RolePermission{
				RoleID:       roleID,
				PermissionID: pid,
			})
		}
		return tx.Create(&rps).Error
	})
}
