package user

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"not null;unique;index:idx_token;size:255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Revoked   bool      `json:"revoked" gorm:"not null;default:false;index"`
	DeviceInfo string   `json:"device_info" gorm:"size:255"` // 设备信息（可选）
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired 检查令牌是否过期
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid 检查令牌是否有效（未过期且未撤销）
func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}

// Revoke 撤销令牌
func (rt *RefreshToken) Revoke(db *gorm.DB) error {
	rt.Revoked = true
	return db.Save(rt).Error
}
