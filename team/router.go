package team

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes 注册团队模块路由
// 基础路径: /api/teams
func RegisterRoutes(g *echo.Group, h *Handler) {
	// 团队相关接口
	g.GET("", h.GetMyTeams)                             // 获取我的团队列表
	g.POST("", h.CreateTeam)                            // 创建新团队
	g.GET("/:id", h.GetTeam)                            // 获取团队详情
	g.GET("/:id/members", h.GetTeamMembers)             // 获取团队成员列表
	g.POST("/:id/members", h.AddMember)                 // 添加成员
	g.DELETE("/:id/members/:uid", h.RemoveMember)       // 移除成员
	g.PUT("/:id/members/:uid", h.UpdateMemberRole)      // 修改成员角色
	g.GET("/:id/intelligences", h.GetTeamIntelligences) // 获取团队情报池
	g.POST("/:id/import", h.ImportIntelligences)        // 批量导入情报到团队
}
