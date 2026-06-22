package repository

import (
	"context"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, keyword string, offset, limit int) ([]model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, keyword string, offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	// 应用数据权限过滤
	if scope, ok := GetDataScope(ctx); ok {
		query = query.Scopes(ApplyScope(scope))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("id ASC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"nickname": user.Nickname,
		"role_id":  user.RoleID,
	}).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}
