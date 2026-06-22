package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-jwt-secret"

func TestGenerateToken_Success(t *testing.T) {
	token, err := GenerateToken(1, 0, "alice", "device-001", 1, testSecret, 1)
	if err != nil {
		t.Fatalf("生成 Token 应成功: %v", err)
	}
	if token == "" {
		t.Fatal("Token 不应为空")
	}
}

func TestParseToken_Valid(t *testing.T) {
	token, _ := GenerateToken(2, 100, "bob", "device-002", 5, testSecret, 1)

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("解析有效 Token 应成功: %v", err)
	}
	if claims.UserID != 2 {
		t.Errorf("UserID 应为 2, 实际: %d", claims.UserID)
	}
	if claims.TenantID != 100 {
		t.Errorf("TenantID 应为 100, 实际: %d", claims.TenantID)
	}
	if claims.Username != "bob" {
		t.Errorf("Username 应为 bob, 实际: %s", claims.Username)
	}
	if claims.DeviceID != "device-002" {
		t.Errorf("DeviceID 应为 device-002, 实际: %s", claims.DeviceID)
	}
	if claims.TokenVersion != 5 {
		t.Errorf("TokenVersion 应为 5, 实际: %d", claims.TokenVersion)
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	token, _ := GenerateToken(1, 0, "alice", "", 0, testSecret, 1)

	_, err := ParseToken(token, "wrong-secret")
	if err != ErrTokenInvalid {
		t.Fatalf("错误签名应返回 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestParseToken_Expired(t *testing.T) {
	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "expired-user",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}
	tokenStr, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("构造过期 Token 失败: %v", err)
	}

	_, err = ParseToken(tokenStr, testSecret)
	if err != ErrTokenExpired {
		t.Fatalf("过期 Token 应返回 ErrTokenExpired, 实际: %v", err)
	}
}

func TestParseToken_GarbageString(t *testing.T) {
	_, err := ParseToken("not.a.valid.jwt", testSecret)
	if err != ErrTokenInvalid {
		t.Fatalf("垃圾字符串应返回 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestParseToken_EmptyString(t *testing.T) {
	_, err := ParseToken("", testSecret)
	if err != ErrTokenInvalid {
		t.Fatalf("空字符串应返回 ErrTokenInvalid, 实际: %v", err)
	}
}

func TestRoundTrip(t *testing.T) {
	token, err := GenerateToken(42, 7, "roundtrip", "dev-x", 99, testSecret, 24)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if claims.UserID != 42 || claims.TenantID != 7 ||
		claims.Username != "roundtrip" || claims.DeviceID != "dev-x" ||
		claims.TokenVersion != 99 {
		t.Fatal("往返校验失败：Claims 字段不匹配")
	}
}
