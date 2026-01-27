package search

import (
	"time"

	"gorm.io/gorm"
)

// SearchResult 搜索结果
type SearchResult struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	AgencyID    uint      `json:"agency_id"`
	AgencyName  string    `json:"agency_name"`
	Keywords    string    `json:"keywords"`
	OriginalURL string    `json:"original_url"`
	PublishDate time.Time `json:"publish_date"`
	CreatedAt   time.Time `json:"created_at"`
	Rating      float64   `json:"rating"`
	IsDuplicate bool      `json:"is_duplicate"` // 是否重复
	DuplicateID uint      `json:"duplicate_id"` // 重复记录的ID
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Q        string `json:"q" validate:"required"`                    // 关键词
	Scope    string `json:"scope" validate:"omitempty"`               // 全网/库内: global, local
	AgencyID uint   `json:"agency_id" validate:"omitempty"`           // 机构ID
	DateFrom string `json:"date_from" validate:"omitempty"`           // 开始日期
	DateTo   string `json:"date_to" validate:"omitempty"`             // 结束日期
	Model    string `json:"model" validate:"omitempty"`               // 模型: basic, advanced, pro
	Limit    int    `json:"limit" validate:"omitempty,min=1,max=100"` // 数量限制
	Page     int    `json:"page" validate:"omitempty,min=1"`          // 页码
}

// CheckDuplicationRequest 查重请求
type CheckDuplicationRequest struct {
	URLs   []string `json:"urls" validate:"omitempty,dive,url"`     // URL列表
	Titles []string `json:"titles" validate:"omitempty,dive,min=1"` // 标题列表
}

// DuplicationResult 查重结果
type DuplicationResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	IsDuplicate bool   `json:"is_duplicate"`
	ExistingID  uint   `json:"existing_id,omitempty"` // 已存在的记录ID
}

// SearchHistory 搜索历史记录
type SearchHistory struct {
	gorm.Model
	UserID      uint      `json:"user_id" gorm:"not null;index;foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Query       string    `json:"query" gorm:"not null;size:500"`
	Scope       string    `json:"scope" gorm:"size:20"`
	ModelType   string    `json:"model_type" gorm:"size:20"` // 避免与 gorm.Model 冲突
	ResultCount int       `json:"result_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// TableName 指定表名
func (SearchHistory) TableName() string {
	return "search_histories"
}
