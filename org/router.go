package org

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/countries", h.GetCountries) // 获取国家列表
	g.GET("/agencies", h.GetAgencies)   // 获取机构列表
}
