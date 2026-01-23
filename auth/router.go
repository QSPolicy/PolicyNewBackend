package auth

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/login", h.Login)    // 用户登录
	g.GET("/session", h.Session) // 检查会话状态
	g.POST("/logout", h.Logout)  // 用户登出
}
