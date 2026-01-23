package auth

import (
	"net/http"
	"policy-backend/utils"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 常量定义
const (
	cookieName   = "jwt"
	cookiePath   = "/"
	cookieMaxAge = 3600 * 24 // 24小时
	userStatusOK = 1
)

// 请求结构体

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdatePasswordRequest 更新密码请求结构体
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// 响应结构体

// AuthResponse 认证响应结构体
type AuthResponse struct {
	Token string `json:"token"`
}

// UserIDResponse 用户ID响应结构体
type UserIDResponse struct {
	ID uint `json:"id"`
}

// UsernameResponse 用户名响应结构体
type UsernameResponse struct {
	Username string `json:"username"`
}

// SessionResponse 会话响应结构体
type SessionResponse struct {
	Active   bool   `json:"active"`
	Username string `json:"username"`
}

// MessageResponse 消息响应结构体
type MessageResponse struct {
	Message string `json:"message"`
}

// Handler 认证处理器
type Handler struct {
	db      *gorm.DB
	jwtUtil *JWTUtil
}

// NewHandler 创建新的认证处理器
func NewHandler(db *gorm.DB, jwtUtil *JWTUtil) *Handler {
	return &Handler{
		db:      db,
		jwtUtil: jwtUtil,
	}
}

// Register 用户注册
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	// 检查用户名是否已存在
	if exists, err := h.userExists(req.Username); err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Database error")
	} else if exists {
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
		Email:        req.Email,
		Status:       userStatusOK,
	}

	if err := h.db.Create(&user).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "User creation failed")
	}

	// 生成JWT令牌
	token, err := h.jwtUtil.GenerateToken(user.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Token generation failed")
	}

	// 设置Cookie
	h.setAuthCookie(c, token)

	return utils.Success(c, AuthResponse{Token: token})
}

// Login 用户登录
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	// 查询用户
	user, err := h.getUserByUsername(req.Username)
	if err != nil {
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
	if user.Status != userStatusOK {
		return utils.Fail(c, http.StatusForbidden, "Account disabled")
	}

	// 生成JWT令牌
	token, err := h.jwtUtil.GenerateToken(user.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Token generation failed")
	}

	// 设置Cookie
	h.setAuthCookie(c, token)

	return utils.Success(c, AuthResponse{Token: token})
}

// GetMyId 获取当前用户ID
func (h *Handler) GetMyId(c echo.Context) error {
	user, err := h.getCurrentUser(c)
	if err != nil {
		return err
	}

	return utils.Success(c, UserIDResponse{ID: user.ID})
}

// GetMyUsername 获取当前用户名
func (h *Handler) GetMyUsername(c echo.Context) error {
	claims, err := h.getTokenClaims(c)
	if err != nil {
		return err
	}

	return utils.Success(c, UsernameResponse{Username: claims.Username})
}

// UpdatePassword 更新密码
func (h *Handler) UpdatePassword(c echo.Context) error {
	var req UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := h.validateRequest(c, &req); err != nil {
		return err
	}

	user, err := h.getCurrentUser(c)
	if err != nil {
		return err
	}

	// 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return utils.Fail(c, http.StatusUnauthorized, "Old password incorrect")
	}

	// 生成新密码哈希
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Password encryption failed")
	}

	user.PasswordHash = string(newHash)
	if err := h.db.Save(&user).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to update password")
	}

	return utils.Success(c, MessageResponse{Message: "password updated"})
}

// Me 获取当前用户信息
func (h *Handler) Me(c echo.Context) error {
	user, err := h.getCurrentUser(c)
	if err != nil {
		return err
	}

	return utils.Success(c, user)
}

// Session 检查会话状态
func (h *Handler) Session(c echo.Context) error {
	claims, err := h.getTokenClaims(c)
	if err != nil {
		return err
	}

	return utils.Success(c, SessionResponse{
		Active:   true,
		Username: claims.Username,
	})
}

// Logout 用户登出
func (h *Handler) Logout(c echo.Context) error {
	h.clearAuthCookie(c)
	return utils.Success(c, MessageResponse{Message: "logged out"})
}

// 辅助方法

// getTokenFromContext 从 cookie 或 Authorization header 中获取 token
func (h *Handler) getTokenFromContext(c echo.Context) string {
	// 先查找 cookie
	if cookie, err := c.Cookie(cookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 再查找 Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	return ""
}

// getTokenClaims 获取token中的声明信息
func (h *Handler) getTokenClaims(c echo.Context) (*JWTClaims, error) {
	token := h.getTokenFromContext(c)
	if token == "" {
		return nil, utils.Fail(c, http.StatusUnauthorized, "Missing token")
	}

	claims, err := h.jwtUtil.ParseToken(token)
	if err != nil {
		return nil, utils.Fail(c, http.StatusUnauthorized, "Invalid token")
	}

	return claims, nil
}

// getCurrentUser 获取当前用户
func (h *Handler) getCurrentUser(c echo.Context) (*User, error) {
	claims, err := h.getTokenClaims(c)
	if err != nil {
		return nil, err
	}

	user, err := h.getUserByUsername(claims.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.Fail(c, http.StatusUnauthorized, "User not found")
		}
		return nil, utils.Error(c, http.StatusInternalServerError, "User lookup failed")
	}

	return user, nil
}

// getUserByUsername 根据用户名查询用户
func (h *Handler) getUserByUsername(username string) (*User, error) {
	var user User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// userExists 检查用户是否存在
func (h *Handler) userExists(username string) (bool, error) {
	var count int64
	if err := h.db.Model(&User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// validateRequest 验证请求数据
func (h *Handler) validateRequest(c echo.Context, req interface{}) error {
	validator := utils.GetValidator(c)
	if validator == nil {
		return utils.Error(c, http.StatusInternalServerError, "Validator not available")
	}
	if err := validator.ValidateStruct(req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, err.Error())
	}
	return nil
}

// setAuthCookie 设置认证Cookie
func (h *Handler) setAuthCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     cookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}

// clearAuthCookie 清除认证Cookie
func (h *Handler) clearAuthCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     cookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
