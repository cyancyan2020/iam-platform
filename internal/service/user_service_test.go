package service

import (
	"context"
	"errors"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository/mocks"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/cyancyan2020/iam-platform/pkg/utils"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret"
const testJWTExpire = 1

func TestRegister_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "newuser" && u.PasswordHash != "" && u.PasswordHash != "secret123"
	})).Return(nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	err := svc.Register(context.Background(), &RegisterRequest{
		Username: "newuser",
		Password: "secret123",
		Nickname: "新用户",
	})
	if err != nil {
		t.Fatalf("正常注册应成功: %v", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "existing").Return(&model.User{ID: 1, Username: "existing"}, nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	err := svc.Register(context.Background(), &RegisterRequest{
		Username: "existing",
		Password: "secret123",
	})
	if !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("期望 ErrUsernameAlreadyExists, 实际: %v", err)
	}
}

func TestRegister_DBErrorOnFind(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "anyone").Return(nil, dbErr)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	err := svc.Register(context.Background(), &RegisterRequest{
		Username: "anyone",
		Password: "secret123",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传数据库错误, 实际: %v", err)
	}
}

func TestRegister_DBErrorOnCreate(t *testing.T) {
	dbErr := errors.New("disk full")
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.Anything).Return(dbErr)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	err := svc.Register(context.Background(), &RegisterRequest{
		Username: "newuser",
		Password: "secret123",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传 Create 错误, 实际: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := utils.HashPassword("correct-password")

	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "testuser").Return(&model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
	}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Incr", mock.Anything, uint64(1)).Return(3, nil)

	svc := NewUserService(repo, tvRepo, testJWTSecret, testJWTExpire)
	resp, err := svc.Login(context.Background(), &LoginRequest{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("登录应成功: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("Token 不应为空")
	}

	claims, err := pkgjwt.ParseToken(resp.Token, testJWTSecret)
	if err != nil {
		t.Fatalf("生成的 Token 应可解析: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "testuser" {
		t.Fatalf("Token Claims 内容不正确: UserID=%d, Username=%s", claims.UserID, claims.Username)
	}
}

func TestLogin_TokenVersionIncrements(t *testing.T) {
	hash, _ := utils.HashPassword("correct-password")

	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "testuser").Return(&model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
	}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Incr", mock.Anything, uint64(1)).Return(5, nil)

	svc := NewUserService(repo, tvRepo, testJWTSecret, testJWTExpire)
	resp, _ := svc.Login(context.Background(), &LoginRequest{
		Username: "testuser",
		Password: "correct-password",
	})

	claims, _ := pkgjwt.ParseToken(resp.Token, testJWTSecret)
	if claims.TokenVersion != 5 {
		t.Fatalf("TokenVersion 应为 Incr 返回值 5, 实际: %d", claims.TokenVersion)
	}
}

func TestLogin_TokenVersionIncrError(t *testing.T) {
	hash, _ := utils.HashPassword("correct-password")

	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "testuser").Return(&model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
	}, nil)

	redisErr := errors.New("redis connection refused")
	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Incr", mock.Anything, uint64(1)).Return(0, redisErr)

	svc := NewUserService(repo, tvRepo, testJWTSecret, testJWTExpire)
	_, err := svc.Login(context.Background(), &LoginRequest{
		Username: "testuser",
		Password: "correct-password",
	})
	if !errors.Is(err, redisErr) {
		t.Fatalf("Redis 异常应透传错误, 实际: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "nobody").Return(nil, gorm.ErrRecordNotFound)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	_, err := svc.Login(context.Background(), &LoginRequest{
		Username: "nobody",
		Password: "whatever",
	})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound, 实际: %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := utils.HashPassword("real-password")

	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "testuser").Return(&model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
	}, nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), testJWTSecret, testJWTExpire)
	_, err := svc.Login(context.Background(), &LoginRequest{
		Username: "testuser",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("期望 ErrInvalidPassword, 实际: %v", err)
	}
}
