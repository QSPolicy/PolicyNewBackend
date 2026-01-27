package user

import (
	"time"

	"gorm.io/gorm"
)

// RefreshTokenService 刷新令牌服务
type RefreshTokenService struct {
	db *gorm.DB
}

// NewRefreshTokenService 创建新的刷新令牌服务
func NewRefreshTokenService(db *gorm.DB) *RefreshTokenService {
	return &RefreshTokenService{db: db}
}

// Create 创建新的刷新令牌
func (s *RefreshTokenService) Create(userID uint, token string, expiresAt time.Time, deviceInfo string) (*RefreshToken, error) {
	refreshToken := &RefreshToken{
		UserID:     userID,
		Token:      token,
		ExpiresAt:  expiresAt,
		Revoked:    false,
		DeviceInfo: deviceInfo,
	}

	if err := s.db.Create(refreshToken).Error; err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// Validate 验证刷新令牌是否有效
func (s *RefreshTokenService) Validate(token string) (*RefreshToken, error) {
	var rt RefreshToken
	if err := s.db.Where("token = ? AND revoked = ? AND expires_at > ?", token, false, time.Now()).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

// Revoke 撤销刷新令牌
func (s *RefreshTokenService) Revoke(token string) error {
	return s.db.Model(&RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

// RevokeAllByUser 撤销指定用户的所有刷新令牌
func (s *RefreshTokenService) RevokeAllByUser(userID uint) error {
	return s.db.Model(&RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

// CleanupExpired 清理过期的刷新令牌
func (s *RefreshTokenService) CleanupExpired() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&RefreshToken{}).Error
}
