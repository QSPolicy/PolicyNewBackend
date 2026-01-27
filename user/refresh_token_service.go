package user

import (
	"time"

	"gorm.io/gorm"
)

// RefreshTokenService Refresh Token 服务
type RefreshTokenService struct {
	db *gorm.DB
}

// NewRefreshTokenService 创建新的 Refresh Token 服务
func NewRefreshTokenService(db *gorm.DB) *RefreshTokenService {
	return &RefreshTokenService{db: db}
}

// Create 创建并保存新的 Refresh Token
func (s *RefreshTokenService) Create(userID uint, token string, expiresIn time.Duration, deviceInfo string) (*RefreshToken, error) {
	rt := &RefreshToken{
		UserID:     userID,
		Token:      token,
		ExpiresAt:  time.Now().Add(expiresIn),
		Revoked:    false,
		DeviceInfo: deviceInfo,
	}

	if err := s.db.Create(rt).Error; err != nil {
		return nil, err
	}

	return rt, nil
}

// FindByToken 根据令牌查找记录
func (s *RefreshTokenService) FindByToken(token string) (*RefreshToken, error) {
	var rt RefreshToken
	if err := s.db.Where("token = ?", token).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

// Validate 验证 Refresh Token 是否有效
func (s *RefreshTokenService) Validate(token string) (*RefreshToken, error) {
	rt, err := s.FindByToken(token)
	if err != nil {
		return nil, err
	}

	if !rt.IsValid() {
		return nil, gorm.ErrRecordNotFound
	}

	return rt, nil
}

// Revoke 撤销指定的 Refresh Token
func (s *RefreshTokenService) Revoke(token string) error {
	rt, err := s.FindByToken(token)
	if err != nil {
		return err
	}

	rt.Revoked = true
	return s.db.Save(rt).Error
}

// RevokeAllByUser 撤销指定用户的所有 Refresh Token
func (s *RefreshTokenService) RevokeAllByUser(userID uint) error {
	return s.db.Model(&RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

// CleanupExpired 清理过期的 Refresh Token
func (s *RefreshTokenService) CleanupExpired() (int64, error) {
	result := s.db.Where("expires_at < ? AND revoked = ?", time.Now(), false).
		Delete(&RefreshToken{})
	return result.RowsAffected, result.Error
}

// GetUserRefreshTokens 获取用户的所有有效 Refresh Token
func (s *RefreshTokenService) GetUserRefreshTokens(userID uint) ([]RefreshToken, error) {
	var tokens []RefreshToken
	err := s.db.Where("user_id = ? AND revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}
