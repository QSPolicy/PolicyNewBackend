package auth

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"not null;unique;size:100" json:"username"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	Nickname     string    `gorm:"size:100" json:"nickname"`
	Status       int       `gorm:"default:1" json:"status"` // 1=active 0=disabled
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate 在创建用户前设置默认值
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Status == 0 {
		u.Status = 1 // 默认启用
	}
	return nil
}
