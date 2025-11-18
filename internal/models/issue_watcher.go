package models

import "time"

// IssueWatcher represents a user watching/subscribing to an issue
type IssueWatcher struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	IssueID   int       `json:"issue_id"`
	CreatedAt time.Time `json:"created_at"`

	// Related entities
	User  *User  `json:"user,omitempty"`
	Issue *Issue `json:"issue,omitempty"`
}
