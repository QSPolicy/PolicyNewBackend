package router

import (
	"policy-backend/auth"
	"policy-backend/config"
	"policy-backend/intelligence"
	"policy-backend/org"
	"policy-backend/user"
	"policy-backend/utils"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func Init(e *echo.Echo, db *gorm.DB, cfg *config.Config) {
	// 1. 统一前缀
	api := e.Group("/api")
	api.Use(middleware.RequestLogger())

	// 2. 注册验证器到Echo
	validator := utils.NewValidator()
	e.Validator = validator
	
	// 将验证器添加到Context中，以便在handler中使用
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("validator", validator)
			return next(c)
		}
	})

	// 3. 初始化各模块并注册
	// Auth 模块
	authH := auth.NewHandler(db, auth.NewJWTUtil(
		cfg.JWTSecretKey,
		time.Duration(cfg.JWTTokenDuration)*time.Hour,
	))
	auth.RegisterRoutes(api.Group("/auth"), authH)

	// User 模块
	userH := user.NewHandler(db)
	user.RegisterRoutes(api.Group("/users"), userH)

	// intelligence 模块
	intelligenceH := intelligence.NewHandler(db)
	intelligence.RegisterPoliciesRoutes(api.Group("/policies"), intelligenceH)
	intelligence.RegisterPolicyRoutes(api.Group("/policy"), intelligenceH)

	// Org 模块
	orgH := org.NewHandler(db)
	org.RegisterRoutes(e.Group("/org"), orgH)
}
