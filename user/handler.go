package user

import (
	"net/http"
	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 常量定义
const (
	userStatusOK = "active"
)

// 请求结构体

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
}

// UpdatePasswordRequest 更新密码请求结构体
type UpdatePasswordRequest struct {
	Username    string `json:"username" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

// 响应结构体

// UserIDResponse 用户ID响应结构体
type UserIDResponse struct {
	ID uint `json:"id"`
}

// UsernameResponse 用户名响应结构体
type UsernameResponse struct {
	Username string `json:"username"`
}

// MessageResponse 消息响应结构体
type MessageResponse struct {
	Message string `json:"message"`
}

// Handler 用户处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建新的用户处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
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

	return utils.Success(c, user)
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

	// 根据用户名获取用户
	user, err := h.getUserByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Fail(c, http.StatusNotFound, "User not found")
		}
		return utils.Error(c, http.StatusInternalServerError, "Database error")
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
