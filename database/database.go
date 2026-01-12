package database

import (
"policy-backend/policy"

"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(databaseURL string) error {
	var err error

	dbPath := "policy.db"

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := DB.AutoMigrate(
&policy.Policy{},
	); err != nil {
		return err
	}

	return nil
}
