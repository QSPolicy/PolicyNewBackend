package auth

import (
	"net/http"
	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse 认证响应结构体
type AuthResponse struct {
	Token string `json:"token"`
}

type Handler struct {
	db      *gorm.DB
	jwtUtil *JWTUtil
}

func NewHandler(db *gorm.DB, jwtUtil *JWTUtil) *Handler {
	return &Handler{
		db:      db,
		jwtUtil: jwtUtil,
	}
}

// Register 用户注册
func (h *Handler) Register(c echo.Context) error {
	// 解析请求
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	// 检查用户名是否已存在
	var existingUser User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return utils.Fail(c, http.StatusConflict, "Username already exists")
	}

	// 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Password encryption failed")
	}

	// 创建用户
	user := User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Nickname:     req.Nickname,
		Status:       1, // 默认启用
	}

	// 保存用户到数据库
	if err := h.db.Create(&user).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "User creation failed")
	}

	// 生成JWT令牌
	token, err := h.jwtUtil.GenerateToken(user.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Token generation failed")
	}

	// 设置Cookie
	c.SetCookie(&http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24, // 24小时
	})

	// 返回响应
	return utils.Success(c, AuthResponse{
		Token: token,
	})
}

// Login 用户登录
func (h *Handler) Login(c echo.Context) error {
	// 解析请求
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	// 查询用户
	var user User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Fail(c, http.StatusUnauthorized, "Invalid username or password")
		}
		return utils.Error(c, http.StatusInternalServerError, "Server internal error")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return utils.Fail(c, http.StatusUnauthorized, "Invalid username or password")
	}

	// 检查用户状态
	if user.Status != 1 {
		return utils.Fail(c, http.StatusForbidden, "Account disabled")
	}

	// 生成JWT令牌
	token, err := h.jwtUtil.GenerateToken(user.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Token generation failed")
	}

	// 设置Cookie
	c.SetCookie(&http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24, // 24小时
	})

	// 返回响应
	return utils.Success(c, AuthResponse{
		Token: token,
	})
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
