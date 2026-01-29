package main

import (
	"log"
	"policy-backend/config"
	"policy-backend/cron"
	"policy-backend/database"
	"policy-backend/router"
	"policy-backend/search"
	"policy-backend/user"

	"github.com/labstack/echo/v4"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// 初始化积分服务
	pointsSvc := user.NewPointsTransactionService(database.DB)

	// 创建搜索处理器（用于定时任务）
	searchH := search.NewHandler(database.DB, pointsSvc)

	// 启动定时任务
	cronJob := cron.NewCronJob(database.DB, searchH)
	cronJob.Start()
	defer cronJob.Stop()

	// 创建Echo实例
	e := echo.New()

	// 注册路由
	router.Init(e, database.DB, cfg)

	// 启动服务器
	if err := e.Start(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
