package intelligence

import (
	"net/http"
	"policy-backend/utils"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// CreateIntelligence 创建新情报
func (h *Handler) CreateIntelligence(c echo.Context) error {
	intelligence := new(Intelligence)
	if err := c.Bind(intelligence); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	// 这里的 AgencyID 和 ContributorID 可能需要从 JWT Token 中获取
	// 假设中间件已经将 userID 放入 context
	// userID := c.Get("user_id").(uint)
	// intelligence.ContributorID = userID

	if err := h.svc.CreateIntelligence(intelligence); err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to create intelligence")
	}

	return utils.Success(c, intelligence)
}

// GetIntelligenceDetail 获取情报详情（含评分）
func (h *Handler) GetIntelligenceDetail(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid ID")
	}

	// 获取当前用户ID (待完善：从token获取)
	// currently hardcoded or assumed to be 0 if not logged in
	userID := uint(0)

	detail, err := h.svc.GetIntelligenceDetail(uint(id), userID)
	if err != nil {
		return utils.Error(c, http.StatusNotFound, "Intelligence not found")
	}

	return utils.Success(c, detail)
}

// ListIntelligences 获取情报列表
func (h *Handler) ListIntelligences(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	keyword := c.QueryParam("keyword")

	data, total, err := h.svc.ListIntelligences(page, pageSize, keyword)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch list")
	}

	return utils.Success(c, map[string]interface{}{
		"list":  data,
		"total": total,
	})
}

// DeleteIntelligence 删除情报
func (h *Handler) DeleteIntelligence(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid ID")
	}

	if err := h.svc.DeleteIntelligence(uint(id)); err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to delete")
	}

	return utils.Success(c, nil)
}

// RateIntelligence 评分
type RatingRequest struct {
	Score int `json:"score"`
}

func (h *Handler) RateIntelligence(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid ID")
	}

	var req RatingRequest
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid body")
	}

	// TODO: Get userID from context
	userID := uint(1)

	if err := h.svc.RateIntelligence(uint(id), userID, req.Score); err != nil {
		return utils.Error(c, http.StatusBadRequest, err.Error())
	}

	return utils.Success(c, nil)
}

// ShareIntelligence 分享情报
func (h *Handler) ShareIntelligence(c echo.Context) error {
	var req ShareRequest
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid body")
	}

	if err := h.svc.ShareIntelligence(req); err != nil {
		return utils.Error(c, http.StatusInternalServerError, err.Error())
	}

	return utils.Success(c, nil)
}
