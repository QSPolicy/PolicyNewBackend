package user

import (
	"time"

	"gorm.io/gorm"
)

// PointsTransactionService 积分交易服务
type PointsTransactionService struct {
	db *gorm.DB
}

// NewPointsTransactionService 创建新的积分交易服务
func NewPointsTransactionService(db *gorm.DB) *PointsTransactionService {
	return &PointsTransactionService{db: db}
}

// Create 创建积分交易记录
func (s *PointsTransactionService) Create(userID uint, amount int64, transactionType string, description string, metadata string) (*PointsTransaction, error) {
	transaction := &PointsTransaction{
		UserID:      userID,
		Amount:      amount,
		Type:        transactionType,
		Description: description,
		Metadata:    metadata,
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetByUser 获取用户的积分交易记录
func (s *PointsTransactionService) GetByUser(userID uint, limit int, offset int) ([]PointsTransaction, error) {
	var transactions []PointsTransaction
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTotalPoints 获取用户的总积分
func (s *PointsTransactionService) GetTotalPoints(userID uint) (int64, error) {
	var total int64
	err := s.db.Model(&PointsTransaction{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetTodayPoints 获取用户今日获得的积分
func (s *PointsTransactionService) GetTodayPoints(userID uint) (int64, error) {
	var total int64
	today := time.Now().Format("2006-01-02")
	err := s.db.Model(&PointsTransaction{}).
		Where("user_id = ? AND DATE(created_at) = ? AND amount > 0", userID, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
