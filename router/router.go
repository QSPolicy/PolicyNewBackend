package router

import (
	"policy-backend/auth"
	"policy-backend/intelligence"
	custommiddleware "policy-backend/middleware"
	"policy-backend/org"
	"policy-backend/search"
	"policy-backend/team"
	"policy-backend/user"
	"policy-backend/utils"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Init 初始化路由，使用auth模块的配置
func Init(e *echo.Echo, db *gorm.DB, authCfg *auth.Config) {
	// 1. 统一前缀
	api := e.Group("/api")
	api.Use(custommiddleware.ZapLogger()) // 使用自定义的 Zap 日志中间件
	// api.Use(middleware.RequestLogger()) // 移除 Echo 默认的 logger

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
		authCfg.JWTSecretKey,
		time.Duration(authCfg.JWTAccessTokenDuration)*time.Minute,
	)

	// 4. 初始化认证中间件
	authMiddleware := custommiddleware.AuthMiddleware(db, jwtUtil)

	// 5. 初始化各模块并注册
	// Auth 模块（不需要认证）
	refreshTokenDuration := time.Duration(authCfg.JWTRefreshTokenDuration) * 24 * time.Hour
	authH := auth.NewHandler(db, jwtUtil, refreshTokenDuration)
	auth.RegisterRoutes(api.Group("/auth"), authH)

	// 初始化积分服务
	pointsSvc := user.NewPointsTransactionService(db)

	// User 模块（需要认证）
	userH := user.NewHandler(db, pointsSvc)
	userGroup := api.Group("/users")
	userGroup.Use(authMiddleware)
	user.RegisterRoutes(userGroup, userH)

	// Search 模块（需要认证）
	searchH := search.NewHandler(db, pointsSvc)
	searchGroup := api.Group("/search")
	searchGroup.Use(authMiddleware)
	search.RegisterRoutes(searchGroup, searchH)

	// Team 模块（需要认证）
	teamH := team.NewHandler(db)
	teamGroup := api.Group("/teams")
	teamGroup.Use(authMiddleware)
	team.RegisterRoutes(teamGroup, teamH)

	// intelligence 模块（需要认证）
	// 使用依赖注入模式
	intelligenceSvc := intelligence.NewService(db)
	intelligenceH := intelligence.NewHandler(intelligenceSvc)

	// 注册 /intelligence 路由组
	intelligenceGroup := api.Group("/intelligence")
	intelligenceGroup.Use(authMiddleware)
	intelligence.RegisterRoutes(intelligenceGroup, intelligenceH)

	// Org 模块（需要认证）
	orgH := org.NewHandler(db)
	orgGroup := e.Group("/org")
	orgGroup.Use(authMiddleware)
	org.RegisterRoutes(orgGroup, orgH)
}
