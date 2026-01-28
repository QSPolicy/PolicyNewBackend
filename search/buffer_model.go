package search

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SearchBuffer 搜索缓冲区模型
// 用于暂存全网检索回来的原始数据
type SearchBuffer struct {
	gorm.Model

	// 会话信息
	SessionID string `gorm:"type:varchar(64);index;comment:检索会话ID" json:"session_id"`
	UserID    uint   `gorm:"index;comment:发起检索的用户ID" json:"user_id"`

	// 原始数据（核心字段）
	RawData json.RawMessage `gorm:"type:json;comment:爬虫抓取的完整原始结构" json:"raw_data"`

	// 预览字段（用于列表快速显示，避免解析JSON）
	PreviewTitle   string    `gorm:"type:varchar(500);comment:预览标题" json:"preview_title"`
	PreviewSource  string    `gorm:"type:varchar(200);comment:预览来源" json:"preview_source"`
	PreviewDate    time.Time `gorm:"comment:预览发布日期" json:"preview_date"`
	PreviewSummary string    `gorm:"type:text;comment:预览摘要" json:"preview_summary"`

	// 查重相关
	DataHash        string `gorm:"type:varchar(64);index;comment:内容哈希（用于快速查重）" json:"data_hash"`
	DuplicateStatus string `gorm:"type:varchar(20);default:'new';comment:查重结果:new新内容,exists已存在" json:"duplicate_status"`

	// 状态管理
	Status     string     `gorm:"type:varchar(20);default:'pending';comment:状态:pending待处理,imported已入库,discarded已丢弃" json:"status"`
	ExpireAt   time.Time  `gorm:"index;comment:过期时间" json:"expire_at"`
	ImportedAt *time.Time `gorm:"comment:入库时间" json:"imported_at,omitempty"`
}

// TableName 指定表名
func (SearchBuffer) TableName() string {
	return "search_buffers"
}

// SearchBufferPreview 用于列表展示的简化结构
type SearchBufferPreview struct {
	ID              uint      `json:"id"`
	SessionID       string    `json:"session_id"`
	PreviewTitle    string    `json:"title"`
	PreviewSource   string    `json:"source"`
	PreviewDate     time.Time `json:"publish_date"`
	PreviewSummary  string    `json:"summary"`
	DuplicateStatus string    `json:"duplicate_status"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToPreview 转换为预览结构
func (b *SearchBuffer) ToPreview() *SearchBufferPreview {
	return &SearchBufferPreview{
		ID:              b.ID,
		SessionID:       b.SessionID,
		PreviewTitle:    b.PreviewTitle,
		PreviewSource:   b.PreviewSource,
		PreviewDate:     b.PreviewDate,
		PreviewSummary:  b.PreviewSummary,
		DuplicateStatus: b.DuplicateStatus,
		Status:          b.Status,
		CreatedAt:       b.CreatedAt,
	}
}

// ImportIntelligenceRequest 导入情报请求
type ImportIntelligenceRequest struct {
	BufferIDs   []uint `json:"buffer_ids" validate:"required,min=1"`
	TargetScope string `json:"target_scope" validate:"required,oneof=mine team"`
	TeamID      uint   `json:"team_id,omitempty"` // 当 target_scope 为 team 时必填
}

// SearchSession 搜索会话
type SearchSession struct {
	ID         string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Query      string    `gorm:"type:varchar(500)" json:"query"`
	Source     string    `gorm:"type:varchar(50)" json:"source"`
	TotalCount int       `json:"total_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (SearchSession) TableName() string {
	return "search_sessions"
}
