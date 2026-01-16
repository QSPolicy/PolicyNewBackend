package database

import (
	"policy-backend/auth"
	"policy-backend/config"
	"policy-backend/policy"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) error {
	var err error
	var db *gorm.DB

	// 根据数据库URL前缀判断数据库类型
	switch {
	case strings.HasPrefix(cfg.DatabaseURL, "mysql://"):
		// MySQL连接格式: mysql://user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		mysqlDSN := strings.TrimPrefix(cfg.DatabaseURL, "mysql://")
		db, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
		if err != nil {
			return err
		}
		// 配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(cfg.MySQLMaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MySQLMaxOpenConns)
	case strings.HasPrefix(cfg.DatabaseURL, "sqlite3://"):
		// SQLite连接格式: sqlite3://path/to/database.db
		dbPath := strings.TrimPrefix(cfg.DatabaseURL, "sqlite3://")
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			return err
		}
	default:
		return gorm.ErrInvalidDB
	}

	DB = db

	// 自动迁移数据库表
	if err := DB.AutoMigrate(
		&policy.Policy{},
		&auth.User{},
	); err != nil {
		return err
	}

	return nil
}
