package repository

import (
	"context"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	HasPermission(ctx context.Context, userID uint64, path, method string) (bool, error)
	FindByID(ctx context.Context, id uint64) (*model.Permission, error)
	List(ctx context.Context) ([]model.Permission, error)
	Create(ctx context.Context, perm *model.Permission) error
	Update(ctx context.Context, perm *model.Permission) error
	Delete(ctx context.Context, id uint64) error
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

func (r *permissionRepository) FindByID(ctx context.Context, id uint64) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.WithContext(ctx).First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

func (r *permissionRepository) Create(ctx context.Context, perm *model.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *permissionRepository) Update(ctx context.Context, perm *model.Permission) error {
	return r.db.WithContext(ctx).Model(perm).Updates(map[string]interface{}{
		"path":      perm.Path,
		"method":    perm.Method,
		"name":      perm.Name,
		"parent_id": perm.ParentID,
	}).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Permission{}, id).Error
}
