package team

import (
	"policy-backend/intelligence"
	"policy-backend/user"
)

// TeamWithMembers 包含成员信息的团队详情
type TeamWithMembers struct {
	*user.Team        `json:",inline"`
	MembersCount      int                    `json:"members_count"`
	IntelligencesCount int                   `json:"intelligences_count"`
	Members           []TeamMemberWithUser  `json:"members"`
}

// TeamMemberWithUser 包含用户信息的团队成员
type TeamMemberWithUser struct {
	*user.TeamMember `json:",inline"`
	User             *user.User           `json:"user"`
}

// TeamIntelligence 团队情报
type TeamIntelligence struct {
	*intelligence.Intelligence `json:",inline"`
	PermissionType string         `json:"permission_type"` // view, edit, admin
}

// CreateTeamRequest 创建团队请求
type CreateTeamRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID uint   `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=admin member"` // admin, member
}

// UpdateMemberRoleRequest 修改成员角色请求
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

// ImportIntelligencesRequest 批量导入情报请求
type ImportIntelligencesRequest struct {
	IntelligenceIDs []uint `json:"intelligence_ids" validate:"required,min=1,dive,min=1"`
}
