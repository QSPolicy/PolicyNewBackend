package handlers

import (
	"net/http"
	"policy-backend/database"
	"policy-backend/models"

	"github.com/labstack/echo/v4"
)

// CreatePolicy 创建新策略
func CreatePolicy(c echo.Context) error {
	policy := new(models.Policy)
	if err := c.Bind(policy); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	if err := database.DB.Create(policy).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create policy"})
	}

	return c.JSON(http.StatusCreated, policy)
}

// GetPolicy 获取单个策略
func GetPolicy(c echo.Context) error {
	id := c.Param("id")
	var policy models.Policy

	if err := database.DB.First(&policy, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Policy not found"})
	}

	return c.JSON(http.StatusOK, policy)
}

// UpdatePolicy 更新策略
func UpdatePolicy(c echo.Context) error {
	id := c.Param("id")
	var policy models.Policy

	// 查找现有策略
	if err := database.DB.First(&policy, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Policy not found"})
	}

	// 绑定更新数据
	if err := c.Bind(&policy); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	if err := database.DB.Save(&policy).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update policy"})
	}

	return c.JSON(http.StatusOK, policy)
}

// DeletePolicy 删除策略
func DeletePolicy(c echo.Context) error {
	id := c.Param("id")

	if err := database.DB.Delete(&models.Policy{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete policy"})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetAllPolicies 获取所有策略
func GetAllPolicies(c echo.Context) error {
	var policies []models.Policy

	if err := database.DB.Find(&policies).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to get policies"})
	}

	return c.JSON(http.StatusOK, policies)
}

// SearchPolicy 搜索政策
func SearchPolicy(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// GetPolicyDetail 获取政策详情
func GetPolicyDetail(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// OrgStats 获取组织统计信息
func OrgStats(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// Home 获取首页数据
func Home(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// PagePolicy 分页获取政策
func PagePolicy(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// ExportCsv 导出CSV
func ExportCsv(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// ManualIngest 手动导入政策
func ManualIngest(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}

// DeleteMyDetail 删除我的详情
func DeleteMyDetail(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "Not implemented yet"})
}
