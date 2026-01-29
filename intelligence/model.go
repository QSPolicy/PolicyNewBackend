package intelligence

import (
	"time"

	"gorm.io/gorm"
)

// Intelligence 情报信息
type Intelligence struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Title         string    `json:"title" gorm:"not null;index"`
	Content       string    `json:"content" gorm:"type:text"`
	AgencyID      uint      `json:"agency_id" gorm:"index"`
	Source        string    `json:"source" gorm:"type:varchar(200)"`
	URL           string    `json:"url" gorm:"type:text"`
	Summary       string    `json:"summary" gorm:"type:text"`
	Keywords      string    `json:"keywords" gorm:"type:text"`
	DataHash      string    `json:"data_hash" gorm:"type:varchar(64);index"`
	ContributorID uint      `json:"contributor_id" gorm:"not null;index"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	TeamID        *uint     `json:"team_id,omitempty" gorm:"index"`
	PublishDate   time.Time `json:"publish_date"`
	Status        string    `json:"status" gorm:"type:varchar(20);default:'temporary'"` // temporary: 临时, official: 正式
}

// 常量定义状态
const (
	StatusTemporary = "temporary"
	StatusOfficial  = "official"
)

// TableName 指定表名
func (Intelligence) TableName() string {
	return "intelligences"
}

// IntelligenceShared 情报共享记录
type IntelligenceShared struct {
	gorm.Model
	IntelligenceID uint   `json:"intelligence_id" gorm:"not null;index;uniqueIndex:idx_shared"`
	TargetUserID   uint   `json:"target_user_id" gorm:"index;uniqueIndex:idx_shared"` // 如果分享给个人
	TargetOrgID    uint   `json:"target_org_id" gorm:"index;uniqueIndex:idx_shared"`  // 如果分享给组织
	SharedType     string `json:"shared_type" gorm:"type:varchar(20);not null"`       // user 或 org
}

// 常量定义分享类型
const (
	ShareTypeUser = "user"
	ShareTypeOrg  = "org"
)

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
