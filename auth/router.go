package auth

import (
"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/register", h.Register)             // 用户注册
	g.GET("/users/Id", h.GetMyId)               // 获取当前用户ID
	g.GET("/users/Name", h.GetMyUsername)       // 获取当前用户名
	g.PUT("/password/change", h.UpdatePassword) // 更新密码
	g.POST("/login", h.Login)                   // 用户登录
	g.GET("/me", h.Me)                          // 获取当前用户信息
	g.GET("/session", h.Session)                // 检查会话状态
	g.POST("/logout", h.Logout)                 // 用户登出
}
