package service

import (
	"context"
	"errors"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newRoleService(t *testing.T) (RoleService, *mocks.RoleRepository, *mocks.PermissionRepository, *mocks.RolePermissionRepository, *mocks.UserRepository) {
	t.Helper()
	roleRepo := mocks.NewRoleRepository(t)
	permRepo := mocks.NewPermissionRepository(t)
	rolePermRepo := mocks.NewRolePermissionRepository(t)
	userRepo := mocks.NewUserRepository(t)
	svc := NewRoleService(roleRepo, permRepo, rolePermRepo, userRepo)
	return svc, roleRepo, permRepo, rolePermRepo, userRepo
}

// ——— AssignRole ———

func TestAssignRole_Success(t *testing.T) {
	svc, roleRepo, _, _, userRepo := newRoleService(t)
	userRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, Username: "test", RoleID: 0}, nil)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, Code: "admin", Name: "管理员"}, nil)
	userRepo.On("UpdateRoleID", mock.Anything, uint64(1), uint64(2)).Return(nil)

	err := svc.AssignRole(context.Background(), 1, &AssignRoleRequest{RoleID: 2})
	if err != nil {
		t.Fatalf("分配角色应成功: %v", err)
	}
}

func TestAssignRole_UserNotFound(t *testing.T) {
	svc, _, _, _, userRepo := newRoleService(t)
	userRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.AssignRole(context.Background(), 99, &AssignRoleRequest{RoleID: 2})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound, 实际: %v", err)
	}
}

func TestAssignRole_RoleNotFound(t *testing.T) {
	svc, roleRepo, _, _, userRepo := newRoleService(t)
	userRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, Username: "test"}, nil)
	roleRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.AssignRole(context.Background(), 1, &AssignRoleRequest{RoleID: 99})
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound, 实际: %v", err)
	}
}

// ——— CreateRole ———

func TestCreateRole_Success(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByCode", mock.Anything, "editor").Return(nil, gorm.ErrRecordNotFound)
	roleRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *model.Role) bool {
		return r.Code == "editor" && r.Name == "编辑者"
	})).Return(nil).Run(func(args mock.Arguments) {
		r := args.Get(1).(*model.Role)
		r.ID = 3
	})

	role, err := svc.CreateRole(context.Background(), &CreateRoleRequest{Code: "editor", Name: "编辑者"})
	if err != nil {
		t.Fatalf("创建角色应成功: %v", err)
	}
	if role.Code != "editor" || role.Name != "编辑者" {
		t.Fatalf("角色内容不正确: Code=%s Name=%s", role.Code, role.Name)
	}
}

func TestCreateRole_CodeAlreadyExists(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByCode", mock.Anything, "admin").Return(&model.Role{ID: 1, Code: "admin"}, nil)

	_, err := svc.CreateRole(context.Background(), &CreateRoleRequest{Code: "admin", Name: "管理员"})
	if !errors.Is(err, ErrRoleCodeAlreadyExists) {
		t.Fatalf("期望 ErrRoleCodeAlreadyExists, 实际: %v", err)
	}
}

func TestCreateRole_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByCode", mock.Anything, "editor").Return(nil, dbErr)

	_, err := svc.CreateRole(context.Background(), &CreateRoleRequest{Code: "editor", Name: "编辑者"})
	if !errors.Is(err, dbErr) {
		t.Fatalf("应透传数据库错误, 实际: %v", err)
	}
}

// ——— ListRoles ———

func TestListRoles_Success(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("List", mock.Anything).Return([]model.Role{
		{ID: 1, Code: "admin", Name: "管理员"},
		{ID: 2, Code: "user", Name: "普通用户"},
	}, nil)

	roles, err := svc.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("获取角色列表应成功: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("期望 2 个角色, 实际: %d", len(roles))
	}
}

// ——— UpdateRole ———

func TestUpdateRole_Success(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, Code: "user", Name: "普通用户"}, nil)
	roleRepo.On("FindByCode", mock.Anything, "member").Return(nil, gorm.ErrRecordNotFound)
	roleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := svc.UpdateRole(context.Background(), 2, &UpdateRoleRequest{Code: "member", Name: "成员"})
	if err != nil {
		t.Fatalf("更新角色应成功: %v", err)
	}
}

func TestUpdateRole_NotFound(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.UpdateRole(context.Background(), 99, &UpdateRoleRequest{Code: "test", Name: "测试"})
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound, 实际: %v", err)
	}
}

func TestUpdateRole_CodeAlreadyExists(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, Code: "user", Name: "普通用户"}, nil)
	roleRepo.On("FindByCode", mock.Anything, "admin").Return(&model.Role{ID: 1, Code: "admin"}, nil)

	err := svc.UpdateRole(context.Background(), 2, &UpdateRoleRequest{Code: "admin", Name: "管理员"})
	if !errors.Is(err, ErrRoleCodeAlreadyExists) {
		t.Fatalf("期望 ErrRoleCodeAlreadyExists, 实际: %v", err)
	}
}

func TestUpdateRole_SameCodeOK(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, Code: "user", Name: "普通用户"}, nil)
	roleRepo.On("FindByCode", mock.Anything, "user").Return(&model.Role{ID: 2, Code: "user"}, nil)
	roleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := svc.UpdateRole(context.Background(), 2, &UpdateRoleRequest{Code: "user", Name: "用户"})
	if err != nil {
		t.Fatalf("相同编码更新自身应成功: %v", err)
	}
}

// ——— DeleteRole ———

func TestDeleteRole_Success(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(3)).Return(&model.Role{ID: 3, Code: "temp"}, nil)
	roleRepo.On("Delete", mock.Anything, uint64(3)).Return(nil)

	err := svc.DeleteRole(context.Background(), 3)
	if err != nil {
		t.Fatalf("删除角色应成功: %v", err)
	}
}

func TestDeleteRole_NotFound(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.DeleteRole(context.Background(), 99)
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound, 实际: %v", err)
	}
}

// ——— ListPermissions ———

func TestListPermissions_Success(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("List", mock.Anything).Return([]model.Permission{
		{ID: 1, Path: "/api/v1/profile", Method: "GET", Name: "个人信息"},
	}, nil)

	perms, err := svc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("获取权限列表应成功: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("期望 1 个权限, 实际: %d", len(perms))
	}
}

// ——— CreatePermission ———

func TestCreatePermission_Success(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Permission) bool {
		return p.Path == "/api/v1/users" && p.Method == "GET" && p.Name == "用户列表"
	})).Return(nil).Run(func(args mock.Arguments) {
		p := args.Get(1).(*model.Permission)
		p.ID = 2
	})

	perm, err := svc.CreatePermission(context.Background(), &CreatePermissionRequest{
		Path:   "/api/v1/users",
		Method: "GET",
		Name:   "用户列表",
	})
	if err != nil {
		t.Fatalf("创建权限应成功: %v", err)
	}
	if perm.Path != "/api/v1/users" {
		t.Fatalf("权限路径不正确: %s", perm.Path)
	}
}

// ——— UpdatePermission ———

func TestUpdatePermission_Success(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.Permission{ID: 1, Path: "/old", Method: "GET", Name: "旧"}, nil)
	permRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	err := svc.UpdatePermission(context.Background(), 1, &UpdatePermissionRequest{
		Path:   "/api/v1/users",
		Method: "GET",
		Name:   "用户列表",
	})
	if err != nil {
		t.Fatalf("更新权限应成功: %v", err)
	}
}

func TestUpdatePermission_NotFound(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.UpdatePermission(context.Background(), 99, &UpdatePermissionRequest{
		Path: "/test", Method: "GET", Name: "测试",
	})
	if !errors.Is(err, ErrPermissionNotFound) {
		t.Fatalf("期望 ErrPermissionNotFound, 实际: %v", err)
	}
}

// ——— DeletePermission ———

func TestDeletePermission_Success(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("FindByID", mock.Anything, uint64(3)).Return(&model.Permission{ID: 3}, nil)
	permRepo.On("Delete", mock.Anything, uint64(3)).Return(nil)

	err := svc.DeletePermission(context.Background(), 3)
	if err != nil {
		t.Fatalf("删除权限应成功: %v", err)
	}
}

func TestDeletePermission_NotFound(t *testing.T) {
	svc, _, permRepo, _, _ := newRoleService(t)
	permRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.DeletePermission(context.Background(), 99)
	if !errors.Is(err, ErrPermissionNotFound) {
		t.Fatalf("期望 ErrPermissionNotFound, 实际: %v", err)
	}
}

// ——— SetRolePermissions ———

func TestSetRolePermissions_Success(t *testing.T) {
	svc, roleRepo, _, rolePermRepo, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.Role{ID: 1, Code: "admin"}, nil)
	rolePermRepo.On("BatchReplace", mock.Anything, uint64(1), []uint64{1, 2, 3}).Return(nil)

	err := svc.SetRolePermissions(context.Background(), 1, &SetRolePermissionsRequest{PermissionIDs: []uint64{1, 2, 3}})
	if err != nil {
		t.Fatalf("设置角色权限应成功: %v", err)
	}
}

func TestSetRolePermissions_RoleNotFound(t *testing.T) {
	svc, roleRepo, _, _, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.SetRolePermissions(context.Background(), 99, &SetRolePermissionsRequest{PermissionIDs: []uint64{1}})
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound, 实际: %v", err)
	}
}

func TestSetRolePermissions_EmptyList(t *testing.T) {
	svc, roleRepo, _, rolePermRepo, _ := newRoleService(t)
	roleRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.Role{ID: 1, Code: "admin"}, nil)
	rolePermRepo.On("BatchReplace", mock.Anything, uint64(1), []uint64{}).Return(nil)

	err := svc.SetRolePermissions(context.Background(), 1, &SetRolePermissionsRequest{PermissionIDs: []uint64{}})
	if err != nil {
		t.Fatalf("清空角色权限应成功: %v", err)
	}
}
