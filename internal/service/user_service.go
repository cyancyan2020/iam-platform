package service

import (
	"context"
	"errors"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	"github.com/cyancyan2020/iam-platform/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrUsernameAlreadyExists = errors.New("用户名已存在")
	ErrUserNotFound          = errors.New("用户不存在")
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=64"`
}

type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(ctx context.Context, req *RegisterRequest) error {
	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return ErrUsernameAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Nickname:     req.Nickname,
	}

	return s.userRepo.Create(ctx, user)
}
