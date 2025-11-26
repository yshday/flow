package models

import "time"

// TasklistItem represents a checklist item within an issue
type TasklistItem struct {
	ID          int        `json:"id"`
	IssueID     int        `json:"issue_id"`
	Content     string     `json:"content"`
	IsCompleted bool       `json:"is_completed"`
	Position    int        `json:"position"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy *int       `json:"completed_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Populated fields
	CompletedByUser *User `json:"completed_by_user,omitempty"`
}

// TasklistProgress represents the completion progress of a tasklist
type TasklistProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
	Percent   int `json:"percent"`
}

// CreateTasklistItemRequest represents a request to create a tasklist item
type CreateTasklistItemRequest struct {
	Content  string `json:"content" validate:"required,min=1,max=500"`
	Position *int   `json:"position,omitempty"`
}

// UpdateTasklistItemRequest represents a request to update a tasklist item
type UpdateTasklistItemRequest struct {
	Content     *string `json:"content,omitempty" validate:"omitempty,min=1,max=500"`
	IsCompleted *bool   `json:"is_completed,omitempty"`
	Position    *int    `json:"position,omitempty"`
}

// ReorderTasklistRequest represents a request to reorder tasklist items
type ReorderTasklistRequest struct {
	ItemIDs []int `json:"item_ids" validate:"required,min=1"`
}

// BulkCreateTasklistRequest represents a request to create multiple tasklist items
type BulkCreateTasklistRequest struct {
	Items []CreateTasklistItemRequest `json:"items" validate:"required,min=1,max=50,dive"`
}
