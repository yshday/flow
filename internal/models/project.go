package models

import "time"

// Project represents a project in the system
type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"` // e.g., "PROJ" for PROJ-1, PROJ-2
	Description *string   `json:"description,omitempty"`
	OwnerID     int       `json:"owner_id"`
	Owner       *User     `json:"owner,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Key         string  `json:"key" validate:"required,min=2,max=10,uppercase"`
	Description *string `json:"description,omitempty"`
	TemplateID  *int    `json:"template_id,omitempty"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ProjectMember represents a member of a project
type ProjectMember struct {
	ProjectID  int       `json:"project_id"`
	UserID     int       `json:"user_id"`
	Role       string    `json:"role"` // owner, admin, member, viewer
	User       *User     `json:"user,omitempty"`
	Project    *Project  `json:"project,omitempty"`
	JoinedAt   time.Time `json:"joined_at"`
	InvitedBy  *int      `json:"invited_by,omitempty"`
}

// ProjectRole represents the role of a user in a project
type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleAdmin  ProjectRole = "admin"
	RoleMember ProjectRole = "member"
	RoleViewer ProjectRole = "viewer"
)

// BoardColumn represents a column in the kanban board
type BoardColumn struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBoardColumnRequest represents the request to create a new board column
type CreateBoardColumnRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Position int    `json:"position" validate:"required,min=0"`
}

// UpdateBoardColumnRequest represents the request to update a board column
type UpdateBoardColumnRequest struct {
	Name     *string `json:"name,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// AddMemberRequest represents the request to add a member to a project
type AddMemberRequest struct {
	UserID int         `json:"user_id" validate:"required"`
	Role   ProjectRole `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	Role ProjectRole `json:"role" validate:"required,oneof=owner admin member viewer"`
}
