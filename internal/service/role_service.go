package service

import (
	"context"
	"errors"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound         = errors.New("角色不存在")
	ErrRoleCodeAlreadyExists = errors.New("角色编码已存在")
	ErrPermissionNotFound   = errors.New("权限不存在")
)

type CreateRoleRequest struct {
	Code string `json:"code" binding:"required,min=2,max=32"`
	Name string `json:"name" binding:"required,min=2,max=64"`
}

type UpdateRoleRequest struct {
	Code string `json:"code" binding:"required,min=2,max=32"`
	Name string `json:"name" binding:"required,min=2,max=64"`
}

type AssignRoleRequest struct {
	RoleID uint64 `json:"role_id" binding:"required"`
}

type CreatePermissionRequest struct {
	Path     string `json:"path" binding:"required,min=1,max=128"`
	Method   string `json:"method" binding:"required,min=1,max=10"`
	Name     string `json:"name" binding:"required,min=1,max=64"`
	ParentID uint64 `json:"parent_id"`
}

type UpdatePermissionRequest struct {
	Path     string `json:"path" binding:"required,min=1,max=128"`
	Method   string `json:"method" binding:"required,min=1,max=10"`
	Name     string `json:"name" binding:"required,min=1,max=64"`
	ParentID uint64 `json:"parent_id"`
}

type SetRolePermissionsRequest struct {
	PermissionIDs []uint64 `json:"permission_ids" binding:"required"`
}

type RoleService interface {
	AssignRole(ctx context.Context, userID uint64, req *AssignRoleRequest) error
	ListRoles(ctx context.Context) ([]model.Role, error)
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error)
	UpdateRole(ctx context.Context, id uint64, req *UpdateRoleRequest) error
	DeleteRole(ctx context.Context, id uint64) error
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*model.Permission, error)
	UpdatePermission(ctx context.Context, id uint64, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, id uint64) error
	SetRolePermissions(ctx context.Context, roleID uint64, req *SetRolePermissionsRequest) error
	GetRolePermissions(ctx context.Context, roleID uint64) ([]uint64, error)
}

type roleService struct {
	roleRepo     repository.RoleRepository
	permRepo     repository.PermissionRepository
	rolePermRepo repository.RolePermissionRepository
	userRepo     repository.UserRepository
}

func NewRoleService(
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	rolePermRepo repository.RolePermissionRepository,
	userRepo repository.UserRepository,
) RoleService {
	return &roleService{
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		rolePermRepo: rolePermRepo,
		userRepo:     userRepo,
	}
}

func (s *roleService) AssignRole(ctx context.Context, userID uint64, req *AssignRoleRequest) error {
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	_, err = s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	return s.userRepo.UpdateRoleID(ctx, userID, req.RoleID)
}

func (s *roleService) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error) {
	_, err := s.roleRepo.FindByCode(ctx, req.Code)
	if err == nil {
		return nil, ErrRoleCodeAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := &model.Role{Code: req.Code, Name: req.Name}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uint64, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	existing, err := s.roleRepo.FindByCode(ctx, req.Code)
	if err == nil && existing.ID != id {
		return ErrRoleCodeAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	role.Code = req.Code
	role.Name = req.Name
	return s.roleRepo.Update(ctx, role)
}

func (s *roleService) DeleteRole(ctx context.Context, id uint64) error {
	_, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return s.roleRepo.Delete(ctx, id)
}

func (s *roleService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.permRepo.List(ctx)
}

func (s *roleService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*model.Permission, error) {
	perm := &model.Permission{
		Path:     req.Path,
		Method:   req.Method,
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	if err := s.permRepo.Create(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *roleService) UpdatePermission(ctx context.Context, id uint64, req *UpdatePermissionRequest) error {
	perm, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return err
	}

	perm.Path = req.Path
	perm.Method = req.Method
	perm.Name = req.Name
	perm.ParentID = req.ParentID
	return s.permRepo.Update(ctx, perm)
}

func (s *roleService) DeletePermission(ctx context.Context, id uint64) error {
	_, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return err
	}
	return s.permRepo.Delete(ctx, id)
}

func (s *roleService) SetRolePermissions(ctx context.Context, roleID uint64, req *SetRolePermissionsRequest) error {
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return s.rolePermRepo.BatchReplace(ctx, roleID, req.PermissionIDs)
}

func (s *roleService) GetRolePermissions(ctx context.Context, roleID uint64) ([]uint64, error) {
	// 仅校验角色是否存在
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	rps, err := s.rolePermRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rps))
	for _, rp := range rps {
		ids = append(ids, rp.PermissionID)
	}
	return ids, nil
}
