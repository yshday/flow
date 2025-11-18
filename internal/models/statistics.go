package models

// ProjectStatistics represents statistics for a project
type ProjectStatistics struct {
	ProjectID int `json:"project_id"`

	// Issue counts by status
	TotalIssues  int `json:"total_issues"`
	OpenIssues   int `json:"open_issues"`
	ClosedIssues int `json:"closed_issues"`

	// Issue counts by priority
	CriticalIssues int `json:"critical_issues"`
	HighIssues     int `json:"high_issues"`
	MediumIssues   int `json:"medium_issues"`
	LowIssues      int `json:"low_issues"`

	// Member statistics
	TotalMembers int `json:"total_members"`

	// Label statistics
	TotalLabels int `json:"total_labels"`

	// Milestone statistics
	TotalMilestones int `json:"total_milestones"`
	OpenMilestones  int `json:"open_milestones"`
}

// IssueStatistics represents statistics for issues
type IssueStatistics struct {
	ProjectID int `json:"project_id"`

	// Time-based statistics
	AverageResolutionDays float64 `json:"average_resolution_days"`

	// Recent activity (last 30 days)
	IssuesCreatedLast30Days int `json:"issues_created_last_30_days"`
	IssuesClosedLast30Days  int `json:"issues_closed_last_30_days"`

	// Distribution
	IssuesByAssignee map[string]int `json:"issues_by_assignee,omitempty"`
}

// UserActivityStatistics represents user activity statistics
type UserActivityStatistics struct {
	UserID int `json:"user_id"`

	// Contribution metrics
	IssuesCreated  int `json:"issues_created"`
	IssuesAssigned int `json:"issues_assigned"`
	IssuesClosed   int `json:"issues_closed"`
	CommentsPosted int `json:"comments_posted"`

	// Recent activity
	RecentIssues   []*Issue   `json:"recent_issues,omitempty"`
	RecentComments []*Comment `json:"recent_comments,omitempty"`
}
