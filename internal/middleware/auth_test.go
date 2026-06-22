package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkgjwt "github.com/cyancyan2020/iam-platform/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-secret-for-middleware"

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testJWTSecret))
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

func generateTestToken(userID uint64, username string) string {
	token, _ := pkgjwt.GenerateToken(userID, 0, username, "", 0, testJWTSecret, 1)
	return token
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupTestRouter()
	token := generateTestToken(1, "testuser")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("有效 Token 期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无 Token 期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无效 Token 期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NoBearerPrefix")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("格式错误请求头期望 401, 实际: %d", w.Code)
	}
}

func TestAuthMiddleware_UserClaimsInContext(t *testing.T) {
	router := setupTestRouter()
	token := generateTestToken(42, "answer")

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
