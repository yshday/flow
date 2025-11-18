package models

import "time"

// IssueSearchRequest represents search criteria for issues
type IssueSearchRequest struct {
	Query      string   `json:"query"`       // Search in title and description
	ProjectID  *int     `json:"project_id"`  // Filter by single project (for backward compatibility)
	ProjectIDs []int    `json:"project_ids"` // Filter by multiple projects (for permission filtering)
	Status     []string `json:"status"`      // Filter by status (open, closed, etc.)
	Priority   []string `json:"priority"`    // Filter by priority (critical, high, medium, low)
	AssigneeID *int     `json:"assignee_id"` // Filter by assignee
	ReporterID *int     `json:"reporter_id"` // Filter by reporter
	LabelIDs   []int    `json:"label_ids"`   // Filter by labels
	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ProjectSearchRequest represents search criteria for projects
type ProjectSearchRequest struct {
	Query  string `json:"query"` // Search in name, description, and key
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// IssueSearchResult represents a single issue in search results
type IssueSearchResult struct {
	ID          int        `json:"id"`
	ProjectID   int        `json:"project_id"`
	ProjectKey  string     `json:"project_key"`
	IssueNumber int        `json:"issue_number"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *int       `json:"assignee_id"`
	ReporterID  int        `json:"reporter_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProjectSearchResult represents a single project in search results
type ProjectSearchResult struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	OwnerID     int       `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchResponse represents paginated search results
type SearchResponse struct {
	Results interface{} `json:"results"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}
