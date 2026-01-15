package database

import (
	"policy-backend/auth"
	"policy-backend/policy"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(databaseURL string) error {
	var err error

	// 移除 sqlite3:// 前缀
	// TODO: 支持更多数据库 并做完善的处理
	dbPath := strings.TrimPrefix(databaseURL, "sqlite3://")

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := DB.AutoMigrate(
		&policy.Policy{},
		&auth.User{},
	); err != nil {
		return err
	}

	return nil
}
