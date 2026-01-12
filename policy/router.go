package policy

import (
	"github.com/labstack/echo/v4"
)

func RegisterPoliciesRoutes(api *echo.Group, h *Handler) {

	{
		api.POST("", h.CreatePolicy)       // 创建策略
		api.GET("/:id", h.GetPolicy)       // 获取单个策略
		api.PUT("/:id", h.UpdatePolicy)    // 更新策略
		api.DELETE("/:id", h.DeletePolicy) // 删除策略
		api.GET("", h.GetAllPolicies)      // 获取所有策略
	}

}
func RegisterPolicyRoutes(api *echo.Group, h *Handler) {
	{
		api.GET("/search", h.SearchPolicy)                  // 搜索政策
		api.GET("/detail/:id", h.GetPolicyDetail)           // 获取政策详情
		api.GET("/org/stats", h.OrgStats)                   // 获取组织统计信息
		api.GET("/home", h.Home)                            // 获取首页数据
		api.GET("/page/:type", h.PagePolicy)                // 分页获取政策
		api.GET("/export/csv", h.ExportCsv)                 // 导出CSV
		api.POST("/ingest/manual", h.ManualIngest)          // 手动导入政策
		api.DELETE("/mydetail/:policyId", h.DeleteMyDetail) // 删除我的详情
	}

}
