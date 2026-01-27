package user

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Email        string `json:"email" gorm:"not null;default:'';size:100"`
	Username     string `json:"username" gorm:"not null;unique;size:100"`
	PasswordHash string `json:"-" gorm:"not null;size:255"`
	Nickname     string `json:"nickname" gorm:"not null;default:'';size:100"`
	Organization string `json:"organization" gorm:"size:100"`
	Points       int    `json:"points" gorm:"default:0"`
	Status       string `json:"status" gorm:"not null;default:'active';size:20"` // active, disabled
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Team 团队表
type Team struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null;size:100"`
	CreatorID uint   `json:"creator_id" gorm:"not null;index"`
}

// TableName 指定表名
func (Team) TableName() string {
	return "teams"
}

// TeamMember 团队成员关系表
type TeamMember struct {
	TeamID    uint   `json:"team_id" gorm:"not null;primaryKey;autoIncrement:false"`
	UserID    uint   `json:"user_id" gorm:"not null;primaryKey;autoIncrement:false"`
	Role      string `json:"role" gorm:"not null;default:'member';size:20"` // admin, member
	CreatedAt uint   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt uint   `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (TeamMember) TableName() string {
	return "team_members"
}

// RefreshToken 刷新令牌模型
type RefreshToken struct {
	gorm.Model
	UserID     uint      `json:"user_id" gorm:"not null;index;foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Token      string    `json:"token" gorm:"not null;unique;index:idx_token;size:255"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null;index"`
	Revoked    bool      `json:"revoked" gorm:"not null;default:false;index"`
	DeviceInfo string    `json:"device_info" gorm:"size:255"` // 设备信息（可选）
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// PointsTransaction 积分交易记录
type PointsTransaction struct {
	gorm.Model
	UserID      uint      `json:"user_id" gorm:"not null;index;foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Amount      int64     `json:"amount" gorm:"not null"`             // 正数为获得，负数为消费
	Type        string    `json:"type" gorm:"not null;size:20;index"` // earn, spend
	Description string    `json:"description" gorm:"size:255"`
	Metadata    string    `json:"metadata" gorm:"type:text"` // JSON格式的额外信息
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// TableName 指定表名
func (PointsTransaction) TableName() string {
	return "points_transactions"
}
