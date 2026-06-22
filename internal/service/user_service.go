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

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=64"`
	RoleID   uint64 `json:"role_id"`
}

type UpdateUserRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=64"`
	RoleID   *uint64 `json:"role_id"`
}

type UserListQuery struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type UserListItem struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	RoleID    uint64 `json:"role_id"`
	RoleName  string `json:"role_name"`
	CreatedAt string `json:"created_at"`
}

type UserListResult struct {
	List  []UserListItem `json:"list"`
	Total int64          `json:"total"`
}

type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	ListUsers(ctx context.Context, query *UserListQuery) (*UserListResult, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) error
	UpdateUser(ctx context.Context, id uint64, req *UpdateUserRequest) error
	DeleteUser(ctx context.Context, id uint64) error
}

type userService struct {
	userRepo         repository.UserRepository
	tokenVersionRepo repository.TokenVersionRepository
	roleRepo         repository.RoleRepository
	jwtSecret        string
	jwtExpireHours   int
}

func NewUserService(userRepo repository.UserRepository, tokenVersionRepo repository.TokenVersionRepository, roleRepo repository.RoleRepository, jwtSecret string, jwtExpireHours int) UserService {
	return &userService{
		userRepo:         userRepo,
		tokenVersionRepo: tokenVersionRepo,
		roleRepo:         roleRepo,
		jwtSecret:        jwtSecret,
		jwtExpireHours:   jwtExpireHours,
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

	version, err := s.tokenVersionRepo.Incr(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	token, err := pkgjwt.GenerateToken(user.ID, 0, user.Username, "", version, s.jwtSecret, s.jwtExpireHours)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

func (s *userService) ListUsers(ctx context.Context, query *UserListQuery) (*UserListResult, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Size < 1 || query.Size > 100 {
		query.Size = 10
	}
	offset := (query.Page - 1) * query.Size

	users, total, err := s.userRepo.List(ctx, query.Keyword, offset, query.Size)
	if err != nil {
		return nil, err
	}

	// 批量获取角色名称映射
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	roleNameMap := make(map[uint64]string)
	for _, r := range roles {
		roleNameMap[r.ID] = r.Name
	}

	list := make([]UserListItem, 0, len(users))
	for _, u := range users {
		list = append(list, UserListItem{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			RoleID:    u.RoleID,
			RoleName:  roleNameMap[u.RoleID],
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	return &UserListResult{List: list, Total: total}, nil
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
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
		RoleID:       req.RoleID,
	}
	return s.userRepo.Create(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, id uint64, req *UpdateUserRequest) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 校验角色存在性（nil=不修改, 非0=需存在）
	if req.RoleID != nil && *req.RoleID != 0 {
		_, err = s.roleRepo.FindByID(ctx, *req.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoleNotFound
			}
			return err
		}
	}

	// 仅更新客户端显式传入的字段
	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.RoleID != nil {
		updates["role_id"] = *req.RoleID
	}

	return s.userRepo.Update(ctx, id, updates)
}

func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(ctx, id)
}
