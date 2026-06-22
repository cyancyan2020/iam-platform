package middleware

import (
	"net/http"

	"github.com/cyancyan2020/iam-platform/internal/repository"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataScopeMiddleware 数据权限中间件：根据用户角色 data_scope 构造查询过滤条件
func DataScopeMiddleware(userRepo repository.UserRepository, roleRepo repository.RoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, ok := c.Get("user")
		if !ok {
			c.Next()
			return
		}

		claims, ok := val.(*pkgjwt.Claims)
		if !ok {
			c.Next()
			return
		}

		// 查询用户获取 role_id
		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			return
		}

		// 查询角色获取 data_scope
		role, err := roleRepo.FindByID(c.Request.Context(), user.RoleID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			return
		}

		scope := repository.DataScope{UserID: claims.UserID}
		switch role.DataScope {
		case "all":
			scope.All = true
		case "self":
			scope.Self = true
		default:
			scope.Self = true
		}

		ctx := repository.SetDataScope(c.Request.Context(), scope)
		c.Request = c.Request.WithContext(ctx)
		c.Set("dataScope", scope)

		c.Next()
	}
}
