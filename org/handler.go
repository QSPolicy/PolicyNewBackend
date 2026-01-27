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

// GetCountries 获取国家列表
func (h *Handler) GetCountries(c echo.Context) error {
	var countries []Country
	if err := h.db.Find(&countries).Error; err != nil {
		return utils.Fail(c, http.StatusInternalServerError, "Failed to fetch countries")
	}

	return utils.Success(c, countries)
}

// GetAgencies 获取机构列表
func (h *Handler) GetAgencies(c echo.Context) error {
	countryID := c.QueryParam("country_id")

	query := h.db.Model(&Agency{}).Preload("Country")

	if countryID != "" {
		query = query.Where("country_id = ?", countryID)
	}

	var agencies []Agency
	if err := query.Find(&agencies).Error; err != nil {
		return utils.Fail(c, http.StatusInternalServerError, "Failed to fetch agencies")
	}

	return utils.Success(c, agencies)
}
