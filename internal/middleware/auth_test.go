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

const testJWTSecret = "test-secret-for-middleware"

func setupTestRouter(t *testing.T, tvRepo *mocks.TokenVersionRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testJWTSecret, tvRepo))
	r.GET("/protected", func(c *gin.Context) {
		claims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "claims missing"})
			return
		}
		c.JSON(http.StatusOK, claims)
	})
	return r
}

func newVersionMockReturning(t *testing.T, v int) *mocks.TokenVersionRepository {
	t.Helper()
	m := mocks.NewTokenVersionRepository(t)
	m.On("Get", mock.Anything, mock.Anything).Return(v, nil)
	return m
}

func generateTestToken(userID uint64, username string, version int) string {
	token, _ := pkgjwt.GenerateToken(userID, 0, username, "", version, testJWTSecret, 1)
	return token
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupTestRouter(t, newVersionMockReturning(t, 0))
	token := generateTestToken(1, "testuser", 0)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("有效 Token 期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router := setupTestRouter(t, mocks.NewTokenVersionRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无 Token 期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupTestRouter(t, mocks.NewTokenVersionRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无效 Token 期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	router := setupTestRouter(t, mocks.NewTokenVersionRepository(t))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NoBearerPrefix")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("格式错误请求头期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_TokenVersionMismatch(t *testing.T) {
	tvRepo := newVersionMockReturning(t, 10)
	router := setupTestRouter(t, tvRepo)

	token := generateTestToken(1, "testuser", 3)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("版本号不匹配期望 401, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_UserClaimsInContext(t *testing.T) {
	router := setupTestRouter(t, newVersionMockReturning(t, 0))
	token := generateTestToken(42, "answer", 0)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际: %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("响应体不应为空，应包含 Claims")
	}
}
