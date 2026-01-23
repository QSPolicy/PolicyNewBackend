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
	userStatusOK = "active"
)

// 请求结构体

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// 响应结构体

// AuthResponse 认证响应结构体
type AuthResponse struct {
	Token string `json:"token"`
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

// Login 用户登录
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 查询用户
	user, err := getUserByUsername(h.db, req.Username)
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
