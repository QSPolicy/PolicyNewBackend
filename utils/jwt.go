package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// 错误定义
	ErrInvalidToken     = errors.New("Invalid token")
	ErrExpiredToken     = errors.New("Token expired")
	ErrInvalidTokenType = errors.New("Invalid token type")
)

// TokenType 令牌类型
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// JWTClaims 定义JWT声明结构
type JWTClaims struct {
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
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

// GenerateAccessToken 生成 Access Token
func (j *JWTUtil) GenerateAccessToken(username string) (string, error) {
	claims := JWTClaims{
		Username: username,
		Type:     TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

// GenerateRefreshToken 生成 Refresh Token
func (j *JWTUtil) GenerateRefreshToken(username string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		Username: username,
		Type:     TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
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

// ParseAccessToken 解析并验证 Access Token
func (j *JWTUtil) ParseAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ParseRefreshToken 解析并验证 Refresh Token
func (j *JWTUtil) ParseRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}
