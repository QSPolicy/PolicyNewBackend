package intelligence

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

// CreatePolicy 创建新策略
func (h *Handler) CreatePolicy(c echo.Context) error {
	policy := new(Intelligence)
	if err := c.Bind(policy); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := h.db.Create(policy).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to create policy")
	}

	return utils.Success(c, policy)
}

// GetPolicy 获取单个策略
func (h *Handler) GetPolicy(c echo.Context) error {
	id := c.Param("id")
	var policy Intelligence

	if err := h.db.First(&policy, id).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "Policy not found")
	}

	return utils.Success(c, policy)
}

// UpdatePolicy 更新策略
func (h *Handler) UpdatePolicy(c echo.Context) error {
	id := c.Param("id")
	var policy Intelligence

	// 查找现有策略
	if err := h.db.First(&policy, id).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "Policy not found")
	}

	// 绑定更新数据
	if err := c.Bind(&policy); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := h.db.Save(&policy).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to update policy")
	}

	return utils.Success(c, policy)
}

// DeletePolicy 删除策略
func (h *Handler) DeletePolicy(c echo.Context) error {
	id := c.Param("id")

	if err := h.db.Delete(&Intelligence{}, id).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to delete policy")
	}

	return utils.Success(c, nil)
}

// GetAllPolicies 获取所有策略
func (h *Handler) GetAllPolicies(c echo.Context) error {
	var policies []Intelligence

	if err := h.db.Find(&policies).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to get policies")
	}

	return utils.Success(c, policies)
}

// SearchPolicy 搜索政策
func (h *Handler) SearchPolicy(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// GetPolicyDetail 获取政策详情
func (h *Handler) GetPolicyDetail(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// OrgStats 获取组织统计信息
func (h *Handler) OrgStats(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// Home 获取首页数据
func (h *Handler) Home(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// PagePolicy 分页获取政策
func (h *Handler) PagePolicy(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// ExportCsv 导出CSV
func (h *Handler) ExportCsv(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// ManualIngest 手动导入政策
func (h *Handler) ManualIngest(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}

// DeleteMyDetail 删除我的详情
func (h *Handler) DeleteMyDetail(c echo.Context) error {
	return utils.Fail(c, http.StatusNotImplemented, "Not implemented yet")
}
