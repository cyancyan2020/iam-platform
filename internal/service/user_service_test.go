package service

import (
	"context"
	"errors"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"gorm.io/gorm"
)

type mockUserRepository struct {
	findByUsername func(ctx context.Context, username string) (*model.User, error)
	create         func(ctx context.Context, user *model.User) error
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return m.findByUsername(ctx, username)
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	return m.create(ctx, user)
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepository{
		findByUsername: func(ctx context.Context, username string) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		create: func(ctx context.Context, user *model.User) error {
			if user.Username != "newuser" {
				t.Errorf("用户名应为 newuser, 实际为 %s", user.Username)
			}
			if user.PasswordHash == "" {
				t.Error("密码哈希不应为空")
			}
			if user.PasswordHash == "secret123" {
				t.Error("密码不应以明文存储")
			}
			return nil
		},
	}

	svc := NewUserService(repo)
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
	repo := &mockUserRepository{
		findByUsername: func(ctx context.Context, username string) (*model.User, error) {
			return &model.User{ID: 1, Username: "existing"}, nil
		},
		create: func(ctx context.Context, user *model.User) error {
			t.Error("重复用户名不应触发 Create")
			return nil
		},
	}

	svc := NewUserService(repo)
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
	repo := &mockUserRepository{
		findByUsername: func(ctx context.Context, username string) (*model.User, error) {
			return nil, dbErr
		},
	}

	svc := NewUserService(repo)
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
	repo := &mockUserRepository{
		findByUsername: func(ctx context.Context, username string) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
		create: func(ctx context.Context, user *model.User) error {
			return dbErr
		},
	}

	svc := NewUserService(repo)
	err := svc.Register(context.Background(), &RegisterRequest{
		Username: "newuser",
		Password: "secret123",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传 Create 错误, 实际: %v", err)
	}
}
