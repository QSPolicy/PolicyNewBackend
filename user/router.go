package user

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/register", h.Register)             // 用户注册
	g.PUT("/password", h.UpdatePassword) // 更新密码
}
