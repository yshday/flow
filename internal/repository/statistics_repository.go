package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
)

// StatisticsRepository handles statistics data operations
type StatisticsRepository struct {
	db *sql.DB
}

// NewStatisticsRepository creates a new statistics repository
func NewStatisticsRepository(db *sql.DB) *StatisticsRepository {
	return &StatisticsRepository{db: db}
}

// GetProjectStatistics retrieves statistics for a project
func (r *StatisticsRepository) GetProjectStatistics(ctx context.Context, projectID int) (*models.ProjectStatistics, error) {
	query := `
		SELECT
			-- Issue counts by status
			COUNT(CASE WHEN i.deleted_at IS NULL THEN 1 END) as total_issues,
			COUNT(CASE WHEN i.status = 'open' AND i.deleted_at IS NULL THEN 1 END) as open_issues,
			COUNT(CASE WHEN i.status = 'closed' AND i.deleted_at IS NULL THEN 1 END) as closed_issues,

			-- Issue counts by priority
			COUNT(CASE WHEN i.priority = 'critical' AND i.deleted_at IS NULL THEN 1 END) as critical_issues,
			COUNT(CASE WHEN i.priority = 'high' AND i.deleted_at IS NULL THEN 1 END) as high_issues,
			COUNT(CASE WHEN i.priority = 'medium' AND i.deleted_at IS NULL THEN 1 END) as medium_issues,
			COUNT(CASE WHEN i.priority = 'low' AND i.deleted_at IS NULL THEN 1 END) as low_issues,

			-- Member count
			(SELECT COUNT(*) FROM project_members WHERE project_id = $1) as total_members,

			-- Label count
			(SELECT COUNT(*) FROM labels WHERE project_id = $1) as total_labels,

			-- Milestone counts
			(SELECT COUNT(*) FROM milestones WHERE project_id = $1) as total_milestones,
			(SELECT COUNT(*) FROM milestones WHERE project_id = $1 AND status = 'open') as open_milestones
		FROM issues i
		WHERE i.project_id = $1
	`

	stats := &models.ProjectStatistics{
		ProjectID: projectID,
	}

	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&stats.TotalIssues,
		&stats.OpenIssues,
		&stats.ClosedIssues,
		&stats.CriticalIssues,
		&stats.HighIssues,
		&stats.MediumIssues,
		&stats.LowIssues,
		&stats.TotalMembers,
		&stats.TotalLabels,
		&stats.TotalMilestones,
		&stats.OpenMilestones,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetIssueStatistics retrieves issue-related statistics for a project
func (r *StatisticsRepository) GetIssueStatistics(ctx context.Context, projectID int) (*models.IssueStatistics, error) {
	query := `
		SELECT
			-- Average resolution time (for closed issues)
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400), 0) as avg_resolution_days,

			-- Issues created in last 30 days
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as created_last_30,

			-- Issues closed in last 30 days
			COUNT(CASE WHEN status = 'closed' AND updated_at >= NOW() - INTERVAL '30 days' THEN 1 END) as closed_last_30
		FROM issues
		WHERE project_id = $1 AND deleted_at IS NULL
	`

	stats := &models.IssueStatistics{
		ProjectID: projectID,
	}

	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&stats.AverageResolutionDays,
		&stats.IssuesCreatedLast30Days,
		&stats.IssuesClosedLast30Days,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUserActivityStatistics retrieves activity statistics for a user
func (r *StatisticsRepository) GetUserActivityStatistics(ctx context.Context, userID int) (*models.UserActivityStatistics, error) {
	query := `
		SELECT
			COUNT(CASE WHEN i.reporter_id = $1 THEN 1 END) as issues_created,
			COUNT(CASE WHEN i.assignee_id = $1 THEN 1 END) as issues_assigned,
			COUNT(CASE WHEN i.assignee_id = $1 AND i.status = 'closed' THEN 1 END) as issues_closed,
			(SELECT COUNT(*) FROM comments WHERE user_id = $1) as comments_posted
		FROM issues i
		WHERE i.deleted_at IS NULL
	`

	stats := &models.UserActivityStatistics{
		UserID: userID,
	}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.IssuesCreated,
		&stats.IssuesAssigned,
		&stats.IssuesClosed,
		&stats.CommentsPosted,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
