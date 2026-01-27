package user

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.PUT("/password", h.UpdatePassword) // 更新密码
	g.GET("/session", h.Session)         // 检查会话状态
	g.POST("/logout", h.Logout)          // 用户登出
}
