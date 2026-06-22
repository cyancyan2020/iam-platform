package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/repository/mocks"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

const permJWTSecret = "test-secret-for-permission"

func setupPermTestRouter(t *testing.T, tvRepo *mocks.TokenVersionRepository, permRepo *mocks.PermissionRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(AuthMiddleware(permJWTSecret, tvRepo))
	r.Use(PermissionCheck(permRepo))
	r.GET("/api/v1/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func makePermToken(userID uint64) string {
	token, _ := pkgjwt.GenerateToken(userID, 0, "testuser", "", 0, permJWTSecret, 1)
	return token
}

func TestPermissionCheck_HasPermission(t *testing.T) {
	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	permRepo := mocks.NewPermissionRepository(t)
	permRepo.On("HasPermission", mock.Anything, uint64(1), "/api/v1/profile", "GET").Return(true, nil)

	router := setupPermTestRouter(t, tvRepo, permRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+makePermToken(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("有权限期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestPermissionCheck_NoPermission(t *testing.T) {
	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	permRepo := mocks.NewPermissionRepository(t)
	permRepo.On("HasPermission", mock.Anything, uint64(2), "/api/v1/profile", "GET").Return(false, nil)

	router := setupPermTestRouter(t, tvRepo, permRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+makePermToken(2))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("无权限期望 403, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestPermissionCheck_Unauthenticated(t *testing.T) {
	tvRepo := mocks.NewTokenVersionRepository(t)
	permRepo := mocks.NewPermissionRepository(t)

	router := setupPermTestRouter(t, tvRepo, permRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未认证期望 401, 实际: %d", w.Code)
	}
}
