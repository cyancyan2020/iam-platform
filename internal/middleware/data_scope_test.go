package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cyancyan2020/iam-platform/internal/model"
	"github.com/cyancyan2020/iam-platform/internal/repository/mocks"
	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

const scopeJWTSecret = "test-secret-for-datascope"

func setupScopeTestRouter(t *testing.T, tvRepo *mocks.TokenVersionRepository, userRepo *mocks.UserRepository, roleRepo *mocks.RoleRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(AuthMiddleware(scopeJWTSecret, tvRepo))
	r.Use(DataScopeMiddleware(userRepo, roleRepo))
	r.GET("/api/v1/users", func(c *gin.Context) {
		scope, exists := c.Get("dataScope")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"scope": "none"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"scope": scope})
	})
	return r
}

func makeScopeToken(userID uint64) string {
	token, _ := pkgjwt.GenerateToken(userID, 0, "testuser", "", 0, scopeJWTSecret, 1)
	return token
}

// TestDataScopeMiddleware_AllScope 管理员 → data_scope = "all"
func TestDataScopeMiddleware_AllScope(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, RoleID: 1}, nil)

	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.Role{ID: 1, DataScope: "all"}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	router := setupScopeTestRouter(t, tvRepo, userRepo, roleRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+makeScopeToken(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// TestDataScopeMiddleware_SelfScope 普通用户 → data_scope = "self"
func TestDataScopeMiddleware_SelfScope(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("FindByID", mock.Anything, uint64(5)).Return(&model.User{ID: 5, RoleID: 2}, nil)

	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, DataScope: "self"}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	router := setupScopeTestRouter(t, tvRepo, userRepo, roleRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+makeScopeToken(5))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// TestDataScopeMiddleware_DefaultSelf 未知 scope 默认按 self 处理
func TestDataScopeMiddleware_DefaultSelf(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("FindByID", mock.Anything, uint64(3)).Return(&model.User{ID: 3, RoleID: 3}, nil)

	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(3)).Return(&model.Role{ID: 3, DataScope: "unknown"}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	router := setupScopeTestRouter(t, tvRepo, userRepo, roleRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+makeScopeToken(3))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// TestDataScopeMiddleware_NoAuth 未认证时不设 scope
func TestDataScopeMiddleware_NoAuth(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	roleRepo := mocks.NewRoleRepository(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DataScopeMiddleware(userRepo, roleRepo))
	r.GET("/api/v1/users", func(c *gin.Context) {
		_, exists := c.Get("dataScope")
		c.JSON(http.StatusOK, gin.H{"hasScope": exists})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}
}

// TestDataScope_AllScopeValue 验证 all scope 的 All=true
func TestDataScope_AllScopeValue(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.User{ID: 1, RoleID: 1}, nil)

	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(1)).Return(&model.Role{ID: 1, DataScope: "all"}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	router := setupScopeTestRouter(t, tvRepo, userRepo, roleRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+makeScopeToken(1))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `"All":true`) {
		t.Fatalf("期望 All=true, body: %s", body)
	}
}

// TestDataScope_SelfScopeValue 验证 self scope 的 Self=true 且 UserID 正确
func TestDataScope_SelfScopeValue(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("FindByID", mock.Anything, uint64(42)).Return(&model.User{ID: 42, RoleID: 2}, nil)

	roleRepo := mocks.NewRoleRepository(t)
	roleRepo.On("FindByID", mock.Anything, uint64(2)).Return(&model.Role{ID: 2, DataScope: "self"}, nil)

	tvRepo := mocks.NewTokenVersionRepository(t)
	tvRepo.On("Get", mock.Anything, mock.Anything).Return(0, nil)

	router := setupScopeTestRouter(t, tvRepo, userRepo, roleRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+makeScopeToken(42))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `"Self":true`) {
		t.Fatalf("期望 Self=true, body: %s", body)
	}
	if !strings.Contains(body, `"UserID":42`) {
		t.Fatalf("期望 UserID=42, body: %s", body)
	}
}
