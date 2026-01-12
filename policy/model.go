package policy

import (
	"time"

	"gorm.io/gorm"
)

// Policy 策略模型
type Policy struct {
	gorm.Model
	Name        string    `json:"name" gorm:"not null;unique"`
	Description string    `json:"description"`
	Status      string    `json:"status" gorm:"default:active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
