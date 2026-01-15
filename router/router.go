package router

import (
	"policy-backend/auth"
	"policy-backend/config"
	"policy-backend/org"
	"policy-backend/policy"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Init(e *echo.Echo, db *gorm.DB, cfg *config.Config) {
	// 1. 统一前缀
	api := e.Group("/api")

	// 2. 初始化各模块并注册
	// Auth 模块
	authH := auth.NewHandler(db, auth.NewJWTUtil(
		cfg.JWTSecretKey,
		time.Duration(cfg.JWTTokenDuration)*time.Hour,
	))
	auth.RegisterRoutes(api.Group("/auth"), authH)

	// Policies 模块
	policyH := policy.NewHandler(db)
	policy.RegisterPoliciesRoutes(api.Group("/policies"), policyH)
	policy.RegisterPolicyRoutes(api.Group("/policy"), policyH)

	// Org 模块
	orgH := org.NewHandler(db)
	org.RegisterRoutes(e.Group("/org"), orgH)
}
