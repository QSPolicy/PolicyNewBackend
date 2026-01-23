package auth

import (
	"policy-backend/user"

	"gorm.io/gorm"
)

// UserAlias 为auth模块提供用户模型的别名
type UserAlias = user.User

// getUserByUsername 从数据库获取用户
func getUserByUsername(db *gorm.DB, username string) (*user.User, error) {
	var u user.User
	if err := db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
