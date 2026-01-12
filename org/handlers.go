package org

import (
	"net/http"
	"policy-backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// SomeOrgs 获取部分组织列表
func (h *Handler) SomeOrgs(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// SomeOrgsMySearch 获取部分组织列表（我的搜索）
func (h *Handler) SomeOrgsMySearch(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetPolicyByOrgIds 根据组织ID获取政策列表
func (h *Handler) GetPolicyByOrgIds(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetPolicyWithRating 根据组织ID获取带评分的政策列表
func (h *Handler) GetPolicyWithRating(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetMyPolicy 根据组织ID获取我的搜索政策列表
func (h *Handler) GetMyPolicy(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}
