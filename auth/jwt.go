package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// 错误定义
	ErrInvalidToken = errors.New("Invalid token")
	ErrExpiredToken = errors.New("Token expired")
)

// JWTClaims 定义JWT声明结构
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTUtil JWT工具结构体
type JWTUtil struct {
	SecretKey     string
	TokenDuration time.Duration
}

// NewJWTUtil 创建新的JWT工具实例
func NewJWTUtil(secretKey string, tokenDuration time.Duration) *JWTUtil {
	return &JWTUtil{
		SecretKey:     secretKey,
		TokenDuration: tokenDuration,
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTUtil) GenerateToken(username string) (string, error) {
	claims := JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

// ParseToken 解析JWT令牌
func (j *JWTUtil) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
