package models

import (
	"time"
)

// ActivityAction represents the type of action performed
type ActivityAction string

const (
	ActionCreated ActivityAction = "created"
	ActionUpdated ActivityAction = "updated"
	ActionDeleted ActivityAction = "deleted"
	ActionMoved   ActivityAction = "moved"
	ActionAdded   ActivityAction = "added"
	ActionRemoved ActivityAction = "removed"
)

// EntityType represents the type of entity the activity is related to
type EntityType string

const (
	EntityTypeIssue   EntityType = "issue"
	EntityTypeComment EntityType = "comment"
	EntityTypeLabel   EntityType = "label"
	EntityTypeMember  EntityType = "member"
	EntityTypeProject EntityType = "project"
	EntityTypeBoard   EntityType = "board"
)

// Activity represents an activity log entry
type Activity struct {
	ID         int       `json:"id"`
	ProjectID  *int      `json:"project_id,omitempty"`
	IssueID    *int      `json:"issue_id,omitempty"`
	UserID     int       `json:"user_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   *int      `json:"entity_id,omitempty"`
	FieldName  *string   `json:"field_name,omitempty"`
	OldValue   *string   `json:"old_value,omitempty"`
	NewValue   *string   `json:"new_value,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	Metadata   *string   `json:"metadata,omitempty"` // JSONB stored as string
	CreatedAt  time.Time `json:"created_at"`

	// Joined fields
	User *User `json:"user,omitempty"`
}

// CreateActivityRequest represents the request to create an activity log entry
type CreateActivityRequest struct {
	ProjectID  *int      `json:"project_id,omitempty"`
	IssueID    *int      `json:"issue_id,omitempty"`
	UserID     int       `json:"user_id" validate:"required"`
	Action     string    `json:"action" validate:"required"`
	EntityType string    `json:"entity_type" validate:"required"`
	EntityID   *int      `json:"entity_id,omitempty"`
	FieldName  *string   `json:"field_name,omitempty"`
	OldValue   *string   `json:"old_value,omitempty"`
	NewValue   *string   `json:"new_value,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	Metadata   *string   `json:"metadata,omitempty"`
}

// ActivityFilter represents filtering options for activities
type ActivityFilter struct {
	ProjectID  *int
	IssueID    *int
	UserID     *int
	Action     *string
	EntityType *string
	Limit      int
	Offset     int
}
