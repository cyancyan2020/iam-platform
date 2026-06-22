package repository

import (
	"context"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	HasPermission(ctx context.Context, userID uint64, path, method string) (bool, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) HasPermission(ctx context.Context, userID uint64, path, method string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM user u
		JOIN role_permission rp ON u.role_id = rp.role_id AND rp.deleted_at IS NULL
		JOIN permission p ON rp.permission_id = p.id AND p.deleted_at IS NULL
		WHERE u.id = ? AND u.deleted_at IS NULL
		  AND p.path = ? AND p.method = ?
	`, userID, path, method).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
