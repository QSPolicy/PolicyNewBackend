package search

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes 注册搜索模块路由
// 基础路径: /api/search
func RegisterRoutes(g *echo.Group, h *Handler) {
	// 搜索相关接口
	g.GET("/global", h.GlobalSearch)           // 全网智能检索
	g.POST("/check-duplication", h.CheckDuplication) // 查重检测
	g.GET("/history", h.GetSearchHistory)      // 获取搜索历史
	g.DELETE("/history", h.ClearSearchHistory) // 清除搜索历史
}
