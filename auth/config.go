package auth

// Config 认证模块配置
type Config struct {
	JWTSecretKey            string `koanf:"jwt_secret_key"`
	JWTAccessTokenDuration  int    `koanf:"jwt_access_token_duration"`  // Access Token 有效期（分钟）
	JWTRefreshTokenDuration int    `koanf:"jwt_refresh_token_duration"` // Refresh Token 有效期（天）
}

// DefaultConfig 返回认证模块的默认配置
func DefaultConfig() Config {
	return Config{
		JWTSecretKey:            "default_jwt_secret_key_change_in_production",
		JWTAccessTokenDuration:  60, // Access Token 默认60分钟
		JWTRefreshTokenDuration: 7,  // Refresh Token 默认7天
	}
}
