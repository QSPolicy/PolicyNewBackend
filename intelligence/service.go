package intelligence

import (
	"errors"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateIntelligence 创建情报 (默认状态为 temporary)
func (s *Service) CreateIntelligence(intelligence *Intelligence) error {
	intelligence.Status = StatusTemporary
	return s.db.Create(intelligence).Error
}

// IntelligenceDetail 包含情报详情和评分
type IntelligenceDetail struct {
	Intelligence
	AvgRating float64 `json:"avg_rating"`
	MyRating  int     `json:"my_rating"`
}

// GetIntelligenceDetail 获取情报详情（包括平均分和当前用户的评分）
func (s *Service) GetIntelligenceDetail(id uint, userID uint) (*IntelligenceDetail, error) {
	var intelligence Intelligence
	if err := s.db.First(&intelligence, id).Error; err != nil {
		return nil, err
	}

	// 获取平均分
	var avgResult struct {
		AvgScore float64
	}
	s.db.Model(&Rating{}).
		Select("AVG(score) as avg_score").
		Where("intelligence_id = ?", id).
		Scan(&avgResult)

	// 获取我的评分
	var myRating Rating
	myScore := 0
	if err := s.db.Where("intelligence_id = ? AND user_id = ?", id, userID).First(&myRating).Error; err == nil {
		myScore = myRating.Score
	}

	return &IntelligenceDetail{
		Intelligence: intelligence,
		AvgRating:    avgResult.AvgScore,
		MyRating:     myScore,
	}, nil
}

// UpdateIntelligenceStatus 更新情报状态 (例如从临时变为正式)
func (s *Service) UpdateIntelligenceStatus(id uint, status string) error {
	return s.db.Model(&Intelligence{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteIntelligence 删除情报
func (s *Service) DeleteIntelligence(id uint) error {
	return s.db.Delete(&Intelligence{}, id).Error
}

// RateIntelligence 对情报进行评分
func (s *Service) RateIntelligence(intelligenceID, userID uint, score int) error {
	if score < 1 || score > 5 {
		return errors.New("score must be between 1 and 5")
	}

	// 检查情报是否存在
	var intelligence Intelligence
	if err := s.db.First(&intelligence, intelligenceID).Error; err != nil {
		return err
	}

	// 检查是否已经评分
	var rating Rating
	err := s.db.Where("intelligence_id = ? AND user_id = ?", intelligenceID, userID).First(&rating).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		rating = Rating{
			IntelligenceID: intelligenceID,
			UserID:         userID,
			Score:          score,
		}
		if err := s.db.Create(&rating).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		rating.Score = score
		if err := s.db.Save(&rating).Error; err != nil {
			return err
		}
	}

	return nil
}

// ShareRequest 分享请求参数
type ShareRequest struct {
	IntelligenceID uint   `json:"intelligence_id"`
	TargetID       uint   `json:"target_id"`   // 用户ID 或 组织ID
	TargetType     string `json:"target_type"` // "user" 或 "org"
}

// ShareIntelligence 分享情报
func (s *Service) ShareIntelligence(req ShareRequest) error {
	share := IntelligenceShared{
		IntelligenceID: req.IntelligenceID,
		SharedType:     req.TargetType,
	}

	switch req.TargetType {
	case ShareTypeUser:
		share.TargetUserID = req.TargetID
	case ShareTypeOrg:
		share.TargetOrgID = req.TargetID
	default:
		return errors.New("invalid target type")
	}

	// 开启事务：创建分享记录 + 更新情报状态为正式
	// 无论之前是 temporary 还是 official，只要被分享了，就可以认为是 official 了
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否重复分享，这里简单处理，如果重复可能报错，前端忽略
		// 为了更稳健，可以由前端或这里查询是否存在
		if err := tx.Create(&share).Error; err != nil {
			// 如果是重复键错误，我们认为分享成功
			// 注意：这取决于数据库驱动的错误码
			return err
		}

		// 将情报状态更新为 official
		if err := tx.Model(&Intelligence{}).
			Where("id = ?", req.IntelligenceID).
			Update("status", StatusOfficial).Error; err != nil {
			return err
		}

		return nil
	})
}

// ListIntelligences 获取情报列表，支持分页和关键词搜索
func (s *Service) ListIntelligences(page, pageSize int, keyword string) ([]Intelligence, int64, error) {
	var intelligences []Intelligence
	var total int64

	db := s.db.Model(&Intelligence{})

	if keyword != "" {
		db = db.Where("title LIKE ? OR summary LIKE ? OR keywords LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Limit(pageSize).
		Offset(offset).
		Order("created_at desc").
		Find(&intelligences).Error

	if err != nil {
		return nil, 0, err
	}

	return intelligences, total, nil
}
