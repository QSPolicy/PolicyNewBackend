package router

import (
	"policy-backend/auth"
	"policy-backend/config"
	"policy-backend/intelligence"
	custommiddleware "policy-backend/middleware"
	"policy-backend/org"
	"policy-backend/search"
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

	// 3. 初始化JWT工具
	jwtUtil := utils.NewJWTUtil(
		cfg.JWTSecretKey,
		time.Duration(cfg.JWTAccessTokenDuration)*time.Minute,
	)

	// 4. 初始化认证中间件
	authMiddleware := custommiddleware.AuthMiddleware(db, jwtUtil)

	// 5. 初始化各模块并注册
	// Auth 模块（不需要认证）
	refreshTokenDuration := time.Duration(cfg.JWTRefreshTokenDuration) * 24 * time.Hour
	authH := auth.NewHandler(db, jwtUtil, refreshTokenDuration)
	auth.RegisterRoutes(api.Group("/auth"), authH)

	// User 模块（需要认证）
	userH := user.NewHandler(db)
	userGroup := api.Group("/users")
	userGroup.Use(authMiddleware)
	user.RegisterRoutes(userGroup, userH)

	// Search 模块（需要认证）
	searchH := search.NewHandler(db)
	searchGroup := api.Group("/search")
	searchGroup.Use(authMiddleware)
	search.RegisterRoutes(searchGroup, searchH)

	// intelligence 模块（需要认证）
	intelligenceH := intelligence.NewHandler(db)
	policiesGroup := api.Group("/policies")
	policiesGroup.Use(authMiddleware)
	intelligence.RegisterPoliciesRoutes(policiesGroup, intelligenceH)

	policyGroup := api.Group("/policy")
	policyGroup.Use(authMiddleware)
	intelligence.RegisterPolicyRoutes(policyGroup, intelligenceH)

	// Org 模块（需要认证）
	orgH := org.NewHandler(db)
	orgGroup := e.Group("/org")
	orgGroup.Use(authMiddleware)
	org.RegisterRoutes(orgGroup, orgH)
}
