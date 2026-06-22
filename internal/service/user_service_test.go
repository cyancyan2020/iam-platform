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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, tvRepo, mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, tvRepo, mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, tvRepo, mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
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

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	_, err := svc.Login(context.Background(), &LoginRequest{
		Username: "testuser",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("期望 ErrInvalidPassword, 实际: %v", err)
	}
}

// ——— ListUsers ———

func TestListUsers_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("List", mock.Anything).Return([]model.Role{{ID: 1, Name: "管理员"}}, nil)
	repo.On("List", mock.Anything, "", 0, 10).Return([]model.User{
		{ID: 1, Username: "admin", Nickname: "Admin", RoleID: 1},
	}, int64(1), nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), roleRepo, testJWTSecret, testJWTExpire)
	result, err := svc.ListUsers(context.Background(), &UserListQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("用户列表应成功: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("期望 1 条, 实际: %d", result.Total)
	}
}

func TestListUsers_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)
	repo.On("List", mock.Anything, "", 0, 10).Return(nil, int64(0), dbErr)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), roleRepo, testJWTSecret, testJWTExpire)
	_, err := svc.ListUsers(context.Background(), &UserListQuery{Page: 1, Size: 10})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传 DB 错误, 实际: %v", err)
	}
}

// ——— CreateUser ———

func TestCreateUser_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "newuser" && u.RoleID == 2
	})).Return(nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "newuser", Password: "secret123", Nickname: "新用户", RoleID: 2,
	})
	if err != nil {
		t.Fatalf("创建用户应成功: %v", err)
	}
}

func TestCreateUser_UsernameExists(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "existing").Return(&model.User{ID: 1}, nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "existing", Password: "secret123",
	})
	if !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("期望 ErrUsernameAlreadyExists, 实际: %v", err)
	}
}

// ——— UpdateUser ———

func ptrStr(s string) *string { return &s }
func ptrUint64(v uint64) *uint64 { return &v }

func TestUpdateUser_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, Nickname: "旧"}, nil)
	repo.On("Update", mock.Anything, uint64(1), mock.MatchedBy(func(m map[string]interface{}) bool {
		return m["nickname"] == "新昵称" && m["role_id"] == uint64(2)
	})).Return(nil)
	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2}, nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), roleRepo, testJWTSecret, testJWTExpire)
	err := svc.UpdateUser(context.Background(), 1, &UpdateUserRequest{Nickname: ptrStr("新昵称"), RoleID: ptrUint64(2)})
	if err != nil {
		t.Fatalf("更新用户应成功: %v", err)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.UpdateUser(context.Background(), 99, &UpdateUserRequest{Nickname: ptrStr("x")})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound, 实际: %v", err)
	}
}

func TestUpdateUser_RoleNotFound(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1}, nil)
	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), roleRepo, testJWTSecret, testJWTExpire)
	err := svc.UpdateUser(context.Background(), 1, &UpdateUserRequest{Nickname: ptrStr("x"), RoleID: ptrUint64(99)})
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound, 实际: %v", err)
	}
}

func TestUpdateUser_OnlyNickname(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1}, nil)
	repo.On("Update", mock.Anything, uint64(1), mock.MatchedBy(func(m map[string]interface{}) bool {
		_, hasRole := m["role_id"]
		return m["nickname"] == "仅改昵称" && !hasRole
	})).Return(nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.UpdateUser(context.Background(), 1, &UpdateUserRequest{Nickname: ptrStr("仅改昵称")})
	if err != nil {
		t.Fatalf("仅更新昵称应成功: %v", err)
	}
}

func TestUpdateUser_RoleIDZeroUnassigns(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, RoleID: 2}, nil)
	repo.On("Update", mock.Anything, uint64(1), mock.MatchedBy(func(m map[string]interface{}) bool {
		return m["role_id"] == uint64(0)
	})).Return(nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.UpdateUser(context.Background(), 1, &UpdateUserRequest{RoleID: ptrUint64(0)})
	if err != nil {
		t.Fatalf("撤销角色应成功: %v", err)
	}
}

func TestListUsers_RoleListFails(t *testing.T) {
	dbErr := errors.New("db error")
	repo := mocks.NewUserRepository(t)
	repo.On("List", mock.Anything, "", 0, 10).Return([]model.User{{ID: 1}}, int64(1), nil)
	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("List", mock.Anything).Return(nil, dbErr)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), roleRepo, testJWTSecret, testJWTExpire)
	_, err := svc.ListUsers(context.Background(), &UserListQuery{Page: 1, Size: 10})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传角色列表错误, 实际: %v", err)
	}
}

func TestCreateUser_DBErrorOnCreate(t *testing.T) {
	dbErr := errors.New("disk full")
	repo := mocks.NewUserRepository(t)
	repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.Anything).Return(dbErr)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "newuser", Password: "secret123",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传 Create 错误, 实际: %v", err)
	}
}

// ——— DeleteUser ———

func TestDeleteUser_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(3)).Return(&model.User{ID: 3}, nil)
	repo.On("Delete", mock.Anything, uint64(3)).Return(nil)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.DeleteUser(context.Background(), 3)
	if err != nil {
		t.Fatalf("删除用户应成功: %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewUserService(repo, mocks.NewTokenVersionRepository(t), mocks.NewRoleRepository(t), testJWTSecret, testJWTExpire)
	err := svc.DeleteUser(context.Background(), 99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound, 实际: %v", err)
	}
}
