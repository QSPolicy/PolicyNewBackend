package user

import (
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Email        string `json:"email" gorm:"not null;size:100"`
	Username     string `json:"username" gorm:"not null;unique;size:100"`
	PasswordHash string `json:"-" gorm:"not null;size:255"`
	Nickname     string `json:"nickname" gorm:"not null;size:100"`
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
