package intelligence

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes 注册路由
func RegisterRoutes(g *echo.Group, h *Handler) {
	// 基础 CRUD
	g.POST("", h.CreateIntelligence)
	g.GET("", h.ListIntelligences)
	g.GET("/:id", h.GetIntelligenceDetail)
	g.DELETE("/:id", h.DeleteIntelligence)

	// 评分
	g.POST("/:id/rate", h.RateIntelligence)

	// 分享
	g.POST("/share", h.ShareIntelligence)
}
