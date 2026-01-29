package search

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"policy-backend/intelligence"
	"policy-backend/user"
	"policy-backend/utils"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler 搜索处理器
type Handler struct {
	db            *gorm.DB
	pointsService *user.PointsTransactionService
}

// NewHandler 创建新的搜索处理器
func NewHandler(db *gorm.DB, pointsService *user.PointsTransactionService) *Handler {
	return &Handler{
		db:            db,
		pointsService: pointsService,
	}
}

// GlobalSearch 全网智能检索
// GET /api/search/global
func (h *Handler) GlobalSearch(c echo.Context) error {
	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 获取当前用户
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	// 1. 生成会话ID
	sessionID := uuid.New().String()

	// 2. 调用搜索（占位实现，实际应调用爬虫服务）
	rawResults := h.performSearch(req)

	// 3. 创建搜索会话记录
	session := SearchSession{
		ID:         sessionID,
		UserID:     currentUser.ID,
		Query:      req.Q,
		Source:     req.Scope,
		TotalCount: len(rawResults),
	}
	if err := h.db.Create(&session).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to create search session")
	}

	// 4. 将结果存入缓冲区
	bufferIDs := []uint{}
	for _, raw := range rawResults {
		bufferID, err := h.saveToBuffer(sessionID, currentUser.ID, raw)
		if err != nil {
			return utils.Error(c, http.StatusInternalServerError, "Failed to save search result to buffer")
		}
		bufferIDs = append(bufferIDs, bufferID)
	}

	// 5. 只返回预览数据给前端
	previews := []*SearchBufferPreview{}
	for _, id := range bufferIDs {
		var buffer SearchBuffer
		if err := h.db.First(&buffer, id).Error; err == nil {
			previews = append(previews, buffer.ToPreview())
		}
	}

	// 6. 扣除积分（如果使用高级模型）
	if req.Model == "advanced" {
		if err := h.deductPoints(currentUser.ID, 10); err != nil {
			return utils.Error(c, http.StatusInternalServerError, "Failed to deduct points")
		}
	}

	return utils.Success(c, map[string]interface{}{
		"session_id": sessionID,
		"query":      req.Q,
		"scope":      req.Scope,
		"model":      req.Model,
		"count":      len(previews),
		"results":    previews,
	})
}

// performSearch 执行搜索（占位实现）
// 实际应调用爬虫或搜索服务
func (h *Handler) performSearch(req SearchRequest) []map[string]interface{} {
	// 占位：返回模拟的搜索结果
	results := []map[string]interface{}{
		{
			"title":        "量子计算在金融领域的应用",
			"source":       "科技日报",
			"url":          "https://example.com/1",
			"content":      "量子计算正逐步应用于金融领域的风险评估和投资组合优化...",
			"publish_date": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			"title":        "最新量子算法研究进展",
			"source":       "学术期刊",
			"url":          "https://example.com/2",
			"content":      "研究团队开发了一种新的量子算法，可显著提升计算效率...",
			"publish_date": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
		},
		{
			"title":        "量子计算机商业化前景分析",
			"source":       "行业报告",
			"url":          "https://example.com/3",
			"content":      "多家科技公司宣布量子计算机商业化计划，预计未来5年将进入实用阶段...",
			"publish_date": time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
		},
	}
	return results
}

// saveToBuffer 将搜索结果存入缓冲区
func (h *Handler) saveToBuffer(sessionID string, userID uint, rawData map[string]interface{}) (uint, error) {
	// 1. 计算内容哈希（用于查重）
	dataHash := h.calculateHash(rawData)

	// 2. 查重检测
	duplicateStatus := "new"
	var existingIntelligence intelligence.Intelligence
	if err := h.db.Where("data_hash = ?", dataHash).First(&existingIntelligence).Error; err == nil {
		duplicateStatus = "exists"
	}

	// 3. 序列化原始数据
	rawJSON, err := json.Marshal(rawData)
	if err != nil {
		return 0, err
	}

	// 4. 提取预览字段
	title, _ := rawData["title"].(string)
	source, _ := rawData["source"].(string)
	summary, _ := rawData["content"].(string)
	publishDateStr, _ := rawData["publish_date"].(string)
	publishDate, _ := time.Parse(time.RFC3339, publishDateStr)

	// 5. 创建缓冲区记录
	buffer := SearchBuffer{
		SessionID:       sessionID,
		UserID:          userID,
		RawData:         rawJSON,
		PreviewTitle:    title,
		PreviewSource:   source,
		PreviewDate:     publishDate,
		PreviewSummary:  summary,
		DataHash:        dataHash,
		DuplicateStatus: duplicateStatus,
		Status:          "pending",
		ExpireAt:        time.Now().Add(24 * time.Hour), // 24小时后过期
	}

	if err := h.db.Create(&buffer).Error; err != nil {
		return 0, err
	}

	return buffer.ID, nil
}

// calculateHash 计算数据哈希
func (h *Handler) calculateHash(data map[string]interface{}) string {
	// 使用 URL 或标题计算哈希
	url, _ := data["url"].(string)
	if url == "" {
		title, _ := data["title"].(string)
		url = title
	}
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// deductPoints 扣除积分
func (h *Handler) deductPoints(userID uint, points int64) error {
	return h.pointsService.AddTransaction(
		userID,
		-points,
		"spend",
		"使用高级搜索模型",
		`{"model": "advanced"}`,
	)
}

// ImportIntelligences 从缓冲区导入情报到正式库
// POST /api/search/import
func (h *Handler) ImportIntelligences(c echo.Context) error {
	var req ImportIntelligenceRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 获取当前用户
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	// 1. 验证缓冲区记录是否属于当前用户
	var buffers []SearchBuffer
	if err := h.db.Where("id IN ? AND user_id = ?", req.BufferIDs, currentUser.ID).Find(&buffers).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to find buffer records")
	}

	if len(buffers) != len(req.BufferIDs) {
		return utils.Fail(c, http.StatusForbidden, "Some buffer records do not belong to you or do not exist")
	}

	// 2. 导入到正式库
	importedIDs := []uint{}
	for _, buffer := range buffers {
		if buffer.Status != "pending" {
			continue // 跳过已处理的记录
		}

		// 解析原始数据
		var rawData map[string]interface{}
		if err := json.Unmarshal(buffer.RawData, &rawData); err != nil {
			continue
		}

		// 创建情报记录
		intelligence := intelligence.Intelligence{
			Title:         buffer.PreviewTitle,
			Content:       rawData["content"].(string),
			Source:        buffer.PreviewSource,
			URL:           rawData["url"].(string),
			Summary:       buffer.PreviewSummary,
			PublishDate:   buffer.PreviewDate,
			DataHash:      buffer.DataHash,
			ContributorID: currentUser.ID,
			UserID:        currentUser.ID,
		}

		// 如果目标是团队
		if req.TargetScope == "team" {
			if req.TeamID == 0 {
				return utils.Fail(c, http.StatusBadRequest, "Team ID is required for team scope")
			}
			intelligence.TeamID = &req.TeamID
		}

		if err := h.db.Create(&intelligence).Error; err != nil {
			return utils.Error(c, http.StatusInternalServerError, "Failed to create intelligence")
		}

		// 更新缓冲区状态
		now := time.Now()
		if err := h.db.Model(&buffer).Updates(map[string]interface{}{
			"status":      "imported",
			"imported_at": now,
		}).Error; err != nil {
			return utils.Error(c, http.StatusInternalServerError, "Failed to update buffer status")
		}

		importedIDs = append(importedIDs, intelligence.ID)
	}

	return utils.Success(c, map[string]interface{}{
		"imported_count":   len(importedIDs),
		"intelligence_ids": importedIDs,
	})
}

// GetSearchSessions 获取用户的搜索会话记录
// GET /api/search/sessions
func (h *Handler) GetSearchSessions(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	var sessions []SearchSession
	if err := h.db.Where("user_id = ?", currentUser.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&sessions).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to get search sessions")
	}

	return utils.Success(c, map[string]interface{}{
		"count":    len(sessions),
		"sessions": sessions,
	})
}

// GetSessionBuffers 获取某个会话的缓冲区数据
// GET /api/search/sessions/:id/buffers
func (h *Handler) GetSessionBuffers(c echo.Context) error {
	sessionID := c.Param("id")

	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	// 验证会话是否属于当前用户
	var session SearchSession
	if err := h.db.Where("id = ? AND user_id = ?", sessionID, currentUser.ID).First(&session).Error; err != nil {
		return utils.Fail(c, http.StatusNotFound, "Session not found")
	}

	var buffers []SearchBuffer
	if err := h.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&buffers).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to get buffer records")
	}

	// 转换为预览格式
	previews := []*SearchBufferPreview{}
	for _, buffer := range buffers {
		previews = append(previews, buffer.ToPreview())
	}

	return utils.Success(c, map[string]interface{}{
		"session": session,
		"count":   len(previews),
		"results": previews,
	})
}

// CheckDuplication 查重检测
// POST /api/search/check-duplication
func (h *Handler) CheckDuplication(c echo.Context) error {
	var req CheckDuplicationRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 支持批量查重
	results := []DuplicationResult{}

	// 检查URL列表
	for _, url := range req.URLs {
		hash := md5.Sum([]byte(url))
		dataHash := hex.EncodeToString(hash[:])

		var existingIntelligence intelligence.Intelligence
		if err := h.db.Where("data_hash = ?", dataHash).First(&existingIntelligence).Error; err == nil {
			results = append(results, DuplicationResult{
				URL:         url,
				IsDuplicate: true,
				ExistingID:  existingIntelligence.ID,
				Title:       existingIntelligence.Title,
			})
		} else {
			results = append(results, DuplicationResult{
				URL:         url,
				IsDuplicate: false,
			})
		}
	}

	// 检查标题列表
	for _, title := range req.Titles {
		hash := md5.Sum([]byte(title))
		dataHash := hex.EncodeToString(hash[:])

		var existingIntelligence intelligence.Intelligence
		if err := h.db.Where("data_hash = ?", dataHash).First(&existingIntelligence).Error; err == nil {
			results = append(results, DuplicationResult{
				Title:       title,
				IsDuplicate: true,
				ExistingID:  existingIntelligence.ID,
			})
		} else {
			results = append(results, DuplicationResult{
				Title:       title,
				IsDuplicate: false,
			})
		}
	}

	return utils.Success(c, map[string]interface{}{
		"count":   len(results),
		"results": results,
	})
}

// CleanupExpiredBuffers 清理过期的缓冲区数据
// 此方法应通过定时任务调用
func (h *Handler) CleanupExpiredBuffers() (int64, error) {
	result := h.db.Where("expire_at < ?", time.Now()).Delete(&SearchBuffer{})
	return result.RowsAffected, result.Error
}
