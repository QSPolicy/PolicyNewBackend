package main

import (
	"log"
	"policy-backend/config"
	"policy-backend/database"
	"policy-backend/routes"

	"github.com/labstack/echo/v4"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建Echo实例
	e := echo.New()

	// 注册路由
	routes.RegisterRoutes(e)

	// 启动服务器
	if err := e.Start(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}