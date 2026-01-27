package auth

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/login", h.Login)       // 用户登录
	g.POST("/register", h.Register) // 用户注册
	g.POST("/refresh", h.Refresh)   // 刷新 Access Token
}
