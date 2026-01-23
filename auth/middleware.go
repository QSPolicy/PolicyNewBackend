package auth

import (
	"context"
	"net/http"

	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// AuthMiddleware 校验 JWT 并把用户信息放到 request context 与 echo.Context
func AuthMiddleware(db *gorm.DB, j *JWTUtil) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 从 cookie 或 Authorization header 中取 token
			token := ""
			if cookie, err := c.Cookie("jwt"); err == nil && cookie.Value != "" {
				token = cookie.Value
			}
			if token == "" {
				auth := c.Request().Header.Get("Authorization")
				if auth != "" {
					// 支持 "Bearer <token>"
					if len(auth) > 7 && (auth[:7] == "Bearer ") {
						token = auth[7:]
					}
				}
			}

			if token == "" {
				return utils.Fail(c, http.StatusUnauthorized, "Missing token")
			}

			claims, err := j.ParseToken(token)
			if err != nil {
				return utils.Fail(c, http.StatusUnauthorized, "Invalid token")
			}

			var user User
			if err := db.Where("username = ?", claims.Username).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return utils.Fail(c, http.StatusUnauthorized, "User not found")
				}
				return utils.Error(c, http.StatusInternalServerError, "Server internal error")
			}

			// 把 user 放到 echo.Context 和 request.Context
			c.Set("user", &user)
			ctx := context.WithValue(c.Request().Context(), ctxUserKey, &user)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
