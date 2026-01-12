package org

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/someorgs", h.SomeOrgs)                         // 获取部分组织列表
	g.GET("/someorgs-mysearch", h.SomeOrgsMySearch)        // 获取部分组织列表（我的搜索）
	g.GET("/getpolicy", h.GetPolicyByOrgIds)               // 根据组织ID获取政策列表
	g.GET("/getpolicy-with-rating", h.GetPolicyWithRating) // 根据组织ID获取带评分的政策列表
	g.GET("/getpolicy-mysearch", h.GetMyPolicy)            // 根据组织ID获取我的搜索政策列表
}
