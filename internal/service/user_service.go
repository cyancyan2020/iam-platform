package service

import (
	"context"
	"errors"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/cyancyan2020/iam-platform/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrUsernameAlreadyExists = errors.New("用户名已存在")
	ErrUserNotFound          = errors.New("用户不存在")
	ErrInvalidPassword       = errors.New("密码错误")
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=64"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

type userService struct {
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpireHours int
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpireHours int) UserService {
	return &userService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
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

func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	token, err := pkgjwt.GenerateToken(user.ID, 0, user.Username, "", 0, s.jwtSecret, s.jwtExpireHours)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}
