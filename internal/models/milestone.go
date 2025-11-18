package models

import (
	"time"
)

// MilestoneStatus represents the status of a milestone
type MilestoneStatus string

const (
	MilestoneStatusOpen   MilestoneStatus = "open"
	MilestoneStatusClosed MilestoneStatus = "closed"
)

// Milestone represents a project milestone
type Milestone struct {
	ID          int             `json:"id"`
	ProjectID   int             `json:"project_id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	DueDate     *time.Time      `json:"due_date,omitempty"`
	Status      MilestoneStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	// Computed fields
	TotalIssues  int `json:"total_issues,omitempty"`
	ClosedIssues int `json:"closed_issues,omitempty"`
	Progress     int `json:"progress,omitempty"` // Percentage (0-100)
}

// CreateMilestoneRequest represents the request to create a milestone
type CreateMilestoneRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description *string    `json:"description,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateMilestoneRequest represents the request to update a milestone
type UpdateMilestoneRequest struct {
	Title       *string           `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string           `json:"description,omitempty"`
	DueDate     *time.Time        `json:"due_date,omitempty"`
	Status      *MilestoneStatus  `json:"status,omitempty" validate:"omitempty,oneof=open closed"`
}
