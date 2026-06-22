package middleware

import (
	"net/http"

	"github.com/cyancyan2020/iam-platform/internal/repository"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func PermissionCheck(permRepo repository.PermissionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
			})
			return
		}

		claims, ok := val.(*pkgjwt.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			return
		}

		hasPerm, err := permRepo.HasPermission(c.Request.Context(), claims.UserID, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			return
		}

		if !hasPerm {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无操作权限",
			})
			return
		}

		c.Next()
	}
}
