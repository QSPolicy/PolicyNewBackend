package main

import (
	"policy-backend/config"
	"policy-backend/cron"
	"policy-backend/database"
	"policy-backend/router"
	"policy-backend/search"
	"policy-backend/user"
	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// 加载所有配置（聚合各模块配置）
	cfg := config.LoadConfig()

	// 初始化日志（注入日志配置）
	utils.InitLogger(&cfg.Log)
	defer zap.L().Sync() // 刷新缓冲

	// 打印已加载的配置（调试信息）
	zap.L().Info("Configuration loaded", zap.Any("config", cfg.GetDebugConfig()))

	// 初始化数据库连接（注入数据库配置）
	if err := database.InitDB(&cfg.Database); err != nil {
		zap.L().Fatal("Failed to connect to database", zap.Error(err))
	}

	// 自动迁移数据库表
	if err := database.AutoMigrate(); err != nil {
		zap.L().Fatal("Failed to auto migrate database", zap.Error(err))
	}

	// 初始化积分服务
	pointsSvc := user.NewPointsTransactionService(database.DB)

	// 创建搜索处理器（用于定时任务）
	searchH := search.NewHandler(database.DB, pointsSvc, cfg.MeiliSearchURL, cfg.MeiliSearchKey)

	// 启动定时任务
	cronJob := cron.NewCronJob(database.DB, searchH)
	cronJob.Start()
	defer cronJob.Stop()

	// 创建Echo实例
	e := echo.New()

	// 注册路由（注入配置）
	router.Init(e, database.DB, cfg)

	// 启动服务器（使用服务器配置）
	if err := e.Start(cfg.Server.ServerAddress); err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}
