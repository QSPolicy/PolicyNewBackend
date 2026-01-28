package search

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes 注册搜索模块路由
// 基础路径: /api/search
func RegisterRoutes(g *echo.Group, h *Handler) {
	// 搜索相关接口
	g.GET("/global", h.GlobalSearch)                 // 全网智能检索
	g.POST("/check-duplication", h.CheckDuplication) // 查重检测

	// 缓冲区相关接口
	g.POST("/import", h.ImportIntelligences)            // 从缓冲区导入情报到正式库
	g.GET("/sessions", h.GetSearchSessions)             // 获取搜索会话记录
	g.GET("/sessions/:id/buffers", h.GetSessionBuffers) // 获取某个会话的缓冲区数据
}
