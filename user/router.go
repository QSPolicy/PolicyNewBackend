package user

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/me", h.GetCurrentUser)       // 获取当前用户信息
	g.PUT("/me", h.UpdateCurrentUser)    // 更新个人资料
	g.GET("/me/points", h.GetUserPoints) // 查询积分余额及流水
	g.PUT("/password", h.UpdatePassword) // 更新密码
	g.GET("/session", h.Session)         // 检查会话状态
	g.POST("/logout", h.Logout)          // 用户登出
}
