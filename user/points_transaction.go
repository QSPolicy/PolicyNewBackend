package user

import (
	"errors"
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

// AddTransaction 添加积分交易（原子操作）
// amount: 变动金额，正数为增加，负数为减少
// txType: 交易类型 (e.g., "earn", "spend", "refund")
func (s *PointsTransactionService) AddTransaction(userID uint, amount int64, txType string, description string, metadata string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 如果是扣减积分，先检查余额
		if amount < 0 {
			var user User
			if err := tx.First(&user, userID).Error; err != nil {
				return err
			}
			// 转换为 int64 进行比较
			if int64(user.Points)+amount < 0 {
				return errors.New("insufficient points")
			}
		}

		// 2. 创建交易记录
		transaction := &PointsTransaction{
			UserID:      userID,
			Amount:      amount,
			Type:        txType,
			Description: description,
			Metadata:    metadata,
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// 3. 更新用户积分
		if err := tx.Model(&User{}).Where("id = ?", userID).
			Update("points", gorm.Expr("points + ?", amount)).Error; err != nil {
			return err
		}

		return nil
	})
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
