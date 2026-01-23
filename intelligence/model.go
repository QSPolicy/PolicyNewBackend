package intelligence

import (
	"time"

	"gorm.io/gorm"
)

// Intelligence 情报信息
type Intelligence struct {
	gorm.Model
	Title         string    `json:"title" gorm:"not null;unique;index"`
	AgencyID      uint      `json:"agency_id" gorm:"not null;index"`
	Summary       string    `json:"summary" gorm:"type:text"`
	Keywords      string    `json:"keywords" gorm:"type:text"`
	OriginalURL   string    `json:"original_url" gorm:"type:text"`
	ContributorID uint      `json:"contributor_id" gorm:"not null;index"`
	PublishDate   time.Time `json:"publish_date"`
}

// TableName 指定表名
func (Intelligence) TableName() string {
	return "intelligences"
}

// IntelligenceShared 情报共享记录
type IntelligenceShared struct {
	gorm.Model
	IntelligenceID uint `json:"intelligence_id" gorm:"not null;index;uniqueIndex:idx_shared"`
	TargetUserID   uint `json:"target_user_id" gorm:"not null;index;uniqueIndex:idx_shared"`
	SharedAt       time.Time `json:"shared_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (IntelligenceShared) TableName() string {
	return "intelligence_shared"
}

// Rating 情报评分
type Rating struct {
	gorm.Model
	IntelligenceID uint `json:"intelligence_id" gorm:"not null;index;uniqueIndex:idx_rating"`
	UserID         uint `json:"user_id" gorm:"not null;index;uniqueIndex:idx_rating"`
	Score          int  `json:"score" gorm:"not null;check:score >= 1 AND score <= 5"` // 评分，范围1-5
}

// TableName 指定表名
func (Rating) TableName() string {
	return "ratings"
}
