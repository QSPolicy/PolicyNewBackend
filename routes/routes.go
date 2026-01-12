package routes

import (
	"policy-backend/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)             // 用户注册
		auth.GET("/users/Id", handlers.GetMyId)               // 获取当前用户ID
		auth.GET("/users/Name", handlers.GetMyUsername)       // 获取当前用户名
		auth.PUT("/password/change", handlers.UpdatePassword) // 更新密码
		auth.POST("/login", handlers.Login)                   // 用户登录
		auth.GET("/me", handlers.Me)                          // 获取当前用户信息
		auth.GET("/session", handlers.Session)                // 检查会话状态
		auth.POST("/logout", handlers.Logout)                 // 用户登出
	}

	policies := api.Group("/policies")
	{
		policies.POST("", handlers.CreatePolicy)       // 创建策略
		policies.GET("/:id", handlers.GetPolicy)       // 获取单个策略
		policies.PUT("/:id", handlers.UpdatePolicy)    // 更新策略
		policies.DELETE("/:id", handlers.DeletePolicy) // 删除策略
		policies.GET("", handlers.GetAllPolicies)      // 获取所有策略
	}

	policy := api.Group("/policy")
	{
		policy.GET("/search", handlers.SearchPolicy)                  // 搜索政策
		policy.GET("/detail/:id", handlers.GetPolicyDetail)           // 获取政策详情
		policy.GET("/org/stats", handlers.OrgStats)                   // 获取组织统计信息
		policy.GET("/home", handlers.Home)                            // 获取首页数据
		policy.GET("/page/:type", handlers.PagePolicy)                // 分页获取政策
		policy.GET("/export/csv", handlers.ExportCsv)                 // 导出CSV
		policy.POST("/ingest/manual", handlers.ManualIngest)          // 手动导入政策
		policy.DELETE("/mydetail/:policyId", handlers.DeleteMyDetail) // 删除我的详情
	}

	org := e.Group("/org")
	{
		org.GET("/someorgs", handlers.SomeOrgs)                         // 获取部分组织列表
		org.GET("/someorgs-mysearch", handlers.SomeOrgsMySearch)        // 获取部分组织列表（我的搜索）
		org.GET("/getpolicy", handlers.GetPolicyByOrgIds)               // 根据组织ID获取政策列表
		org.GET("/getpolicy-with-rating", handlers.GetPolicyWithRating) // 根据组织ID获取带评分的政策列表
		org.GET("/getpolicy-mysearch", handlers.GetMyPolicy)            // 根据组织ID获取我的搜索政策列表
	}
}
