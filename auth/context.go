package auth

import (
	"errors"
	"policy-backend/user"

	"github.com/labstack/echo/v4"
)

type ctxKeyType string

const ctxUserKey ctxKeyType = "auth_user"

// GetUser 从 echo.Context 中读取已存放的用户
func GetUser(c echo.Context) (*user.User, bool) {
	if v := c.Get("user"); v != nil {
		if u, ok := v.(*user.User); ok {
			return u, true
		}
	}
	// 也尝试从 request context 中读取
	if v := c.Request().Context().Value(ctxUserKey); v != nil {
		if u, ok := v.(*user.User); ok {
			return u, true
		}
	}
	return nil, false
}

// MustGetUser 获取用户或返回错误
func MustGetUser(c echo.Context) (*user.User, error) {
	if u, ok := GetUser(c); ok {
		return u, nil
	}
	return nil, errors.New("user not found in context")
}

// GetUserID 从 context 中获取用户 ID
func GetUserID(c echo.Context) (uint, bool) {
	if u, ok := GetUser(c); ok {
		return u.ID, true
	}
	return 0, false
}

// GetUsername 从 context 中获取用户名
func GetUsername(c echo.Context) (string, bool) {
	if u, ok := GetUser(c); ok {
		return u.Username, true
	}
	return "", false
}
