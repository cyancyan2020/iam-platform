package handler

import (
	"net/http"
	"strconv"

	"github.com/cyancyan2020/iam-platform/internal/service"
	pkgl "github.com/cyancyan2020/iam-platform/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) AssignRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户 ID"})
		return
	}

	var req service.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	if err := h.roleService.AssignRole(c.Request.Context(), userID, &req); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		case service.ErrRoleNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		default:
			pkgl.Error("AssignRole", zap.Uint64("userID", userID), zap.Uint64("roleID", req.RoleID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色分配成功"})
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.roleService.ListRoles(c.Request.Context())
	if err != nil {
		pkgl.Error("ListRoles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": roles})
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrRoleCodeAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		pkgl.Error("CreateRole", zap.String("code", req.Code), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 201, "data": role})
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色 ID"})
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	if err := h.roleService.UpdateRole(c.Request.Context(), id, &req); err != nil {
		switch err {
		case service.ErrRoleNotFound:
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		case service.ErrRoleCodeAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
		default:
			pkgl.Error("UpdateRole", zap.Uint64("id", id), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色更新成功"})
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色 ID"})
		return
	}

	if err := h.roleService.DeleteRole(c.Request.Context(), id); err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		pkgl.Error("DeleteRole", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色删除成功"})
}

func (h *RoleHandler) SetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色 ID"})
		return
	}

	var req service.SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	if err := h.roleService.SetRolePermissions(c.Request.Context(), roleID, &req); err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		pkgl.Error("SetRolePermissions", zap.Uint64("roleID", roleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "权限分配成功"})
}

func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色 ID"})
		return
	}

	permIDs, err := h.roleService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		pkgl.Error("GetRolePermissions", zap.Uint64("roleID", roleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": permIDs})
}
