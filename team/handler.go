package team

import (
	"net/http"
	"policy-backend/user"
	"policy-backend/utils"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler 团队处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建新的团队处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// GetMyTeams 获取我的团队列表
// GET /api/teams
func (h *Handler) GetMyTeams(c echo.Context) error {
	// 获取当前用户
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	// 查询用户所属的所有团队
	var teamMembers []user.TeamMember
	if err := h.db.Where("user_id = ?", currentUser.ID).Find(&teamMembers).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch team memberships")
	}

	// 获取团队详情
	var teams []TeamWithMembers
	for _, tm := range teamMembers {
		var t user.Team
		if err := h.db.First(&t, tm.TeamID).Error; err != nil {
			continue
		}

		// 统计成员数
		var membersCount int64
		h.db.Model(&user.TeamMember{}).Where("team_id = ?", t.ID).Count(&membersCount)

		teams = append(teams, TeamWithMembers{
			Team:               &t,
			MembersCount:       int(membersCount),
			IntelligencesCount: 0, // TODO: 后续实现
		})
	}

	return utils.Success(c, teams)
}

// CreateTeam 创建新团队
// POST /api/teams
func (h *Handler) CreateTeam(c echo.Context) error {
	// 获取当前用户
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	var req CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 创建团队
	team := &user.Team{
		Name:      req.Name,
		CreatorID: currentUser.ID,
	}

	if err := h.db.Create(team).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to create team")
	}

	// 将创建者添加为管理员
	teamMember := &user.TeamMember{
		TeamID: team.ID,
		UserID: currentUser.ID,
		Role:   "admin",
	}

	if err := h.db.Create(teamMember).Error; err != nil {
		// 回滚团队创建
		h.db.Delete(team)
		return utils.Error(c, http.StatusInternalServerError, "Failed to add creator to team")
	}

	return utils.Success(c, team)
}

// GetTeam 获取团队详情
// GET /api/teams/:id
func (h *Handler) GetTeam(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	// 检查用户是否是团队成员
	if err := h.checkTeamMembership(c, uint(teamID)); err != nil {
		return err
	}

	// 获取团队信息
	var team user.Team
	if err := h.db.First(&team, teamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Fail(c, http.StatusNotFound, "Team not found")
		}
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch team")
	}

	// 获取团队成员
	var teamMembers []user.TeamMember
	if err := h.db.Where("team_id = ?", teamID).Find(&teamMembers).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch team members")
	}

	// 获取成员的用户信息
	var members []TeamMemberWithUser
	for _, tm := range teamMembers {
		var u user.User
		if err := h.db.First(&u, tm.UserID).Error; err == nil {
			members = append(members, TeamMemberWithUser{
				TeamMember: &tm,
				User:       &u,
			})
		}
	}

	result := TeamWithMembers{
		Team:               &team,
		MembersCount:       len(members),
		IntelligencesCount: 0, // TODO: 后续实现
		Members:            members,
	}

	return utils.Success(c, result)
}

// GetTeamMembers 获取团队成员列表
// GET /api/teams/:id/members
func (h *Handler) GetTeamMembers(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	// 检查用户是否是团队成员
	if err := h.checkTeamMembership(c, uint(teamID)); err != nil {
		return err
	}

	// 获取团队成员
	var teamMembers []user.TeamMember
	if err := h.db.Where("team_id = ?", teamID).Find(&teamMembers).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch team members")
	}

	// 获取成员的用户信息
	var members []TeamMemberWithUser
	for _, tm := range teamMembers {
		var u user.User
		if err := h.db.First(&u, tm.UserID).Error; err == nil {
			members = append(members, TeamMemberWithUser{
				TeamMember: &tm,
				User:       &u,
			})
		}
	}

	return utils.Success(c, members)
}

// AddMember 添加成员
// POST /api/teams/:id/members
func (h *Handler) AddMember(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	// 检查用户是否是团队管理员
	if err := h.checkTeamAdmin(c, uint(teamID)); err != nil {
		return err
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 检查用户是否存在
	var u user.User
	if err := h.db.First(&u, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Fail(c, http.StatusNotFound, "User not found")
		}
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch user")
	}

	// 检查用户是否已经是团队成员
	var existingMember user.TeamMember
	if err := h.db.Where("team_id = ? AND user_id = ?", teamID, req.UserID).First(&existingMember).Error; err == nil {
		return utils.Fail(c, http.StatusConflict, "User is already a team member")
	}

	// 添加成员
	teamMember := &user.TeamMember{
		TeamID: uint(teamID),
		UserID: req.UserID,
		Role:   req.Role,
	}

	if err := h.db.Create(teamMember).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to add member to team")
	}

	return utils.Success(c, teamMember)
}

// RemoveMember 移除成员
// DELETE /api/teams/:id/members/:uid
func (h *Handler) RemoveMember(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	userID, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid user ID")
	}

	// 检查用户是否是团队管理员
	if err := h.checkTeamAdmin(c, uint(teamID)); err != nil {
		return err
	}

	// 检查要移除的用户是否是团队成员
	var teamMember user.TeamMember
	if err := h.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&teamMember).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Fail(c, http.StatusNotFound, "User is not a team member")
		}
		return utils.Error(c, http.StatusInternalServerError, "Failed to fetch team member")
	}

	// 移除成员
	if err := h.db.Delete(&teamMember).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to remove member from team")
	}

	return utils.Success(c, map[string]string{
		"message": "Member removed successfully",
	})
}

// UpdateMemberRole 修改成员角色
// PUT /api/teams/:id/members/:uid
func (h *Handler) UpdateMemberRole(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	userID, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid user ID")
	}

	// 检查用户是否是团队管理员
	if err := h.checkTeamAdmin(c, uint(teamID)); err != nil {
		return err
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// 更新角色
	if err := h.db.Model(&user.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", req.Role).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to update member role")
	}

	return utils.Success(c, map[string]string{
		"message": "Member role updated successfully",
	})
}

// GetTeamIntelligences 获取团队情报池
// GET /api/teams/:id/intelligences
func (h *Handler) GetTeamIntelligences(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	// 检查用户是否是团队成员
	if err := h.checkTeamMembership(c, uint(teamID)); err != nil {
		return err
	}

	// TODO: 实现团队情报池查询
	// 需要查询 permissions 表中 subject_type = 'team' 且 subject_id = teamID 的资源

	// 占位返回
	intelligences := []TeamIntelligence{}

	return utils.Success(c, map[string]interface{}{
		"count":         len(intelligences),
		"intelligences": intelligences,
		"message":       "团队情报池功能待实现",
	})
}

// ImportIntelligences 批量导入情报到团队
// POST /api/teams/:id/import
func (h *Handler) ImportIntelligences(c echo.Context) error {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid team ID")
	}

	// 检查用户是否是团队管理员
	if err := h.checkTeamAdmin(c, uint(teamID)); err != nil {
		return err
	}

	var req ImportIntelligencesRequest
	if err := c.Bind(&req); err != nil {
		return utils.Fail(c, http.StatusBadRequest, "Invalid parameters")
	}

	if err := utils.ValidateRequest(c, &req); err != nil {
		return err
	}

	// TODO: 实现批量导入功能
	// 1. 检查情报是否存在
	// 2. 为团队添加权限
	// 3. 处理重复导入

	return utils.Success(c, map[string]interface{}{
		"count":   len(req.IntelligenceIDs),
		"message": "批量导入功能待实现",
	})
}

// checkTeamMembership 检查用户是否是团队成员
func (h *Handler) checkTeamMembership(c echo.Context, teamID uint) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	var count int64
	if err := h.db.Model(&user.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, currentUser.ID).
		Count(&count).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to check team membership")
	}

	if count == 0 {
		return utils.Fail(c, http.StatusForbidden, "You are not a member of this team")
	}

	return nil
}

// checkTeamAdmin 检查用户是否是团队管理员
func (h *Handler) checkTeamAdmin(c echo.Context, teamID uint) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return utils.Fail(c, http.StatusUnauthorized, "User not authenticated")
	}

	var count int64
	if err := h.db.Model(&user.TeamMember{}).
		Where("team_id = ? AND user_id = ? AND role = ?", teamID, currentUser.ID, "admin").
		Count(&count).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to check team admin")
	}

	if count == 0 {
		return utils.Fail(c, http.StatusForbidden, "Only team admins can perform this action")
	}

	return nil
}
