package middleware

import (
	"net/http"
	"strings"

	"github.com/cyancyan2020/iam-platform/internal/repository"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string, tokenVersionRepo repository.TokenVersionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少 Authorization 头",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization 格式错误",
			})
			return
		}

		claims, err := pkgjwt.ParseToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token 无效或已过期",
			})
			return
		}

		currentVersion, err := tokenVersionRepo.Get(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务器内部错误",
			})
			return
		}

		if claims.TokenVersion < currentVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token 已失效（账号在其他设备登录）",
			})
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}
