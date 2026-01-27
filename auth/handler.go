package auth

import (
	"net/http"
	"policy-backend/user"
	"policy-backend/utils"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 常量定义
const (
	cookieName         = "jwt"
	refreshCookieName  = "refresh_token"
	cookiePath         = "/"
	accessTokenMaxAge  = 3600 * 24     // 24小时（实际以配置为准）
	refreshTokenMaxAge = 3600 * 24 * 7 // 7天
	userStatusOK       = "active"
)

// 请求结构体

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest 刷新Token请求结构体
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
}

// 响应结构体

// AuthResponse 认证响应结构体（双Token）
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// MessageResponse 消息响应结构体
type MessageResponse struct {
	Message string `json:"message"`
}

// Handler 认证处理器
type Handler struct {
	db                   *gorm.DB
	jwtUtil              *utils.JWTUtil
	refreshTokenService  *user.RefreshTokenService
	refreshTokenDuration time.Duration
}

// NewHandler 创建新的认证处理器
func NewHandler(db *gorm.DB, jwtUtil *utils.JWTUtil, refreshTokenDuration time.Duration) *Handler {
	return &Handler{
		db:                   db,
		jwtUtil:              jwtUtil,
		refreshTokenService:  user.NewRefreshTokenService(db),
		refreshTokenDuration: refreshTokenDuration,
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

	// 生成 Access Token
	accessToken, err := h.jwtUtil.GenerateAccessToken(user.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Access token generation failed")
	}

	// 生成 Refresh Token
	refreshToken, err := h.jwtUtil.GenerateRefreshToken(user.Username, h.refreshTokenDuration)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Refresh token generation failed")
	}

	// 保存 Refresh Token 到数据库
	deviceInfo := c.Request().UserAgent()
	if _, err := h.refreshTokenService.Create(user.ID, refreshToken, h.refreshTokenDuration, deviceInfo); err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to save refresh token")
	}

	// 设置 Access Token Cookie
	h.setAuthCookie(c, accessToken)

	// 设置 Refresh Token Cookie（HttpOnly，安全）
	h.setRefreshCookie(c, refreshToken)

	return utils.Success(c, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh 刷新 Access Token
func (h *Handler) Refresh(c echo.Context) error {
	// 从 Cookie 或请求体中获取 Refresh Token
	refreshToken := ""

	// 尝试从 Cookie 获取
	if cookie, err := c.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	}

	// 如果 Cookie 中没有，尝试从请求体获取
	if refreshToken == "" {
		var req RefreshRequest
		if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		return utils.Fail(c, http.StatusBadRequest, "Refresh token is required")
	}

	// 验证 Refresh Token 的签名
	_, err := h.jwtUtil.ParseRefreshToken(refreshToken)
	if err != nil {
		return utils.Fail(c, http.StatusUnauthorized, "Invalid refresh token")
	}

	// 检查数据库中的 Refresh Token 状态
	rtRecord, err := h.refreshTokenService.Validate(refreshToken)
	if err != nil {
		return utils.Fail(c, http.StatusUnauthorized, "Refresh token is invalid or expired")
	}

	// 获取用户信息
	var u user.User
	if err := h.db.Where("id = ?", rtRecord.UserID).First(&u).Error; err != nil {
		return utils.Fail(c, http.StatusUnauthorized, "User not found")
	}

	// 检查用户状态
	if u.Status != userStatusOK {
		return utils.Fail(c, http.StatusForbidden, "Account disabled")
	}

	// 生成新的 Access Token
	newAccessToken, err := h.jwtUtil.GenerateAccessToken(u.Username)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to generate access token")
	}

	// 更新 Access Token Cookie
	h.setAuthCookie(c, newAccessToken)

	return utils.Success(c, AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken, // Refresh Token 不变
	})
}

// Register 用户注册
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 检查用户名是否已存在
	var count int64
	if err := h.db.Model(&user.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Database error")
	} else if count > 0 {
		return utils.Fail(c, http.StatusConflict, "Username already exists")
	}

	// 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Password encryption failed")
	}

	// 创建用户
	newUser := user.User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Nickname:     req.Nickname,
		Email:        req.Email,
		Status:       userStatusOK,
	}

	if err := h.db.Create(&newUser).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "User creation failed")
	}

	return utils.Success(c, newUser)
}

// setAuthCookie 设置认证Cookie
func (h *Handler) setAuthCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     cookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   accessTokenMaxAge,
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

// setRefreshCookie 设置 Refresh Token Cookie
func (h *Handler) setRefreshCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   false, // 生产环境建议设为 true（需要HTTPS）
		SameSite: http.SameSiteStrictMode,
		MaxAge:   refreshTokenMaxAge,
	})
}

// clearRefreshCookie 清除 Refresh Token Cookie
func (h *Handler) clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     cookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// getRefreshTokenFromContext 从 Cookie 中获取 Refresh Token
func (h *Handler) getRefreshTokenFromContext(c echo.Context) string {
	if cookie, err := c.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}
