package models

import "time"

// IssueStatus represents the status of an issue
type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusClosed     IssueStatus = "closed"
)

// IssuePriority represents the priority of an issue
type IssuePriority string

const (
	PriorityLow    IssuePriority = "low"
	PriorityMedium IssuePriority = "medium"
	PriorityHigh   IssuePriority = "high"
	PriorityUrgent IssuePriority = "urgent"
)

// Issue represents an issue in the system
type Issue struct {
	ID             int            `json:"id"`
	ProjectID      int            `json:"project_id"`
	IssueNumber    int            `json:"issue_number"`
	Title          string         `json:"title"`
	Description    *string        `json:"description,omitempty"`
	DescriptionHTML *string       `json:"description_html,omitempty"` // Rendered HTML from markdown
	Status         IssueStatus    `json:"status"`
	ColumnID       *int           `json:"column_id,omitempty"`
	ColumnPosition *int           `json:"column_position,omitempty"`
	Priority       IssuePriority  `json:"priority"`
	AssigneeID     *int           `json:"assignee_id,omitempty"`
	ReporterID     int            `json:"reporter_id"`
	MilestoneID    *int           `json:"milestone_id,omitempty"`
	Version        int            `json:"version"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty"`
	IsPinned       bool           `json:"is_pinned"`
	PinnedAt       *time.Time     `json:"pinned_at,omitempty"`
	PinnedByUserID *int           `json:"pinned_by_user_id,omitempty"`

	// Related entities (for joins)
	Assignee  *User      `json:"assignee,omitempty"`
	Reporter  *User      `json:"reporter,omitempty"`
	Project   *Project   `json:"project,omitempty"`
	Labels    []*Label   `json:"labels,omitempty"`
}

// CreateIssueRequest represents the request to create a new issue
type CreateIssueRequest struct {
	Title       string         `json:"title" validate:"required,min=1,max=500"`
	Description *string        `json:"description,omitempty"`
	Priority    *IssuePriority `json:"priority,omitempty"`
	AssigneeID  *int           `json:"assignee_id,omitempty"`
	ColumnID    *int           `json:"column_id,omitempty"`
	MilestoneID *int           `json:"milestone_id,omitempty"`
	LabelIDs    []int          `json:"label_ids,omitempty"`
}

// UpdateIssueRequest represents the request to update an issue
type UpdateIssueRequest struct {
	Title       *string        `json:"title,omitempty"`
	Description *string        `json:"description,omitempty"`
	Status      *IssueStatus   `json:"status,omitempty"`
	Priority    *IssuePriority `json:"priority,omitempty"`
	AssigneeID  *int           `json:"assignee_id,omitempty"`
	MilestoneID *int           `json:"milestone_id,omitempty"`
	Version     *int           `json:"version,omitempty"` // For optimistic locking
}

// MoveIssueRequest represents the request to move an issue to a different column
type MoveIssueRequest struct {
	ColumnID int          `json:"column_id" validate:"required"`
	Position *int         `json:"position,omitempty"`
	Version  int          `json:"version" validate:"required"` // For optimistic locking
	Status   *IssueStatus `json:"status,omitempty"`            // Optional: auto-update status based on column
}

// IssueFilter represents filters for listing issues
type IssueFilter struct {
	ProjectID   int
	Status      *IssueStatus
	Priority    *IssuePriority
	AssigneeID  *int
	ReporterID  *int
	MilestoneID *int
	LabelIDs    []int
	Search      string
	Limit       int
	Offset      int
}

// Label represents a label that can be attached to issues
type Label struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"` // Hex color code
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateLabelRequest represents the request to create a new label
type CreateLabelRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=50"`
	Color       string  `json:"color" validate:"required,hexcolor"`
	Description *string `json:"description,omitempty"`
}

// UpdateLabelRequest represents the request to update a label
type UpdateLabelRequest struct {
	Name        *string `json:"name,omitempty"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Comment represents a comment on an issue
type Comment struct {
	ID          int       `json:"id"`
	IssueID     int       `json:"issue_id"`
	UserID      int       `json:"user_id"`
	Content     string    `json:"content"`
	ContentHTML *string   `json:"content_html,omitempty"` // Rendered HTML from markdown
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Related entities
	User *User `json:"user,omitempty"`
}

// CreateCommentRequest represents the request to create a new comment
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}
