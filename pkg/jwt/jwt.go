package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("Token 已过期")
	ErrTokenInvalid = errors.New("Token 无效")
)

type Claims struct {
	UserID       uint64 `json:"user_id"`
	TenantID     uint64 `json:"tenant_id"`
	Username     string `json:"username"`
	DeviceID     string `json:"device_id"`
	TokenVersion int    `json:"token_version"`
	jwtlib.RegisteredClaims
}

func GenerateToken(userID, tenantID uint64, username, deviceID string, tokenVersion int, secret string, expireHours int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Username:     username,
		DeviceID:     deviceID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
