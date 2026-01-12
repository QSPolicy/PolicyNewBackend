package auth

import (
	"net/http"
	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Register 用户注册
func (h *Handler) Register(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// Login 用户登录
func (h *Handler) Login(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetMyId 获取当前用户ID
func (h *Handler) GetMyId(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetMyUsername 获取当前用户名
func (h *Handler) GetMyUsername(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// UpdatePassword 更新密码
func (h *Handler) UpdatePassword(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// Me 获取当前用户信息
func (h *Handler) Me(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// Session 检查会话状态
func (h *Handler) Session(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// Logout 用户登出
func (h *Handler) Logout(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}
