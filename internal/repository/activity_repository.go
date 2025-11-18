package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yourusername/issue-tracker/internal/models"
)

// ActivityRepository handles activity data access
type ActivityRepository struct {
	db *sql.DB
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create creates a new activity log entry
func (r *ActivityRepository) Create(ctx context.Context, activity *models.Activity) (*models.Activity, error) {
	query := `
		INSERT INTO activities (
			project_id, issue_id, user_id, action, entity_type, entity_id,
			field_name, old_value, new_value, ip_address, user_agent, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		activity.ProjectID,
		activity.IssueID,
		activity.UserID,
		activity.Action,
		activity.EntityType,
		activity.EntityID,
		activity.FieldName,
		activity.OldValue,
		activity.NewValue,
		activity.IPAddress,
		activity.UserAgent,
		activity.Metadata,
	).Scan(&activity.ID, &activity.CreatedAt)

	if err != nil {
		return nil, err
	}

	return activity, nil
}

// ListByProjectID lists all activities for a project
func (r *ActivityRepository) ListByProjectID(ctx context.Context, projectID int, limit, offset int) ([]*models.Activity, error) {
	query := `
		SELECT
			a.id, a.project_id, a.issue_id, a.user_id, a.action, a.entity_type,
			a.entity_id, a.field_name, a.old_value, a.new_value, a.ip_address,
			a.user_agent, a.metadata, a.created_at,
			u.id, u.email, u.username, u.created_at, u.updated_at
		FROM activities a
		JOIN users u ON a.user_id = u.id
		WHERE a.project_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanActivities(rows)
}

// ListByIssueID lists all activities for an issue
func (r *ActivityRepository) ListByIssueID(ctx context.Context, issueID int, limit, offset int) ([]*models.Activity, error) {
	query := `
		SELECT
			a.id, a.project_id, a.issue_id, a.user_id, a.action, a.entity_type,
			a.entity_id, a.field_name, a.old_value, a.new_value, a.ip_address,
			a.user_agent, a.metadata, a.created_at,
			u.id, u.email, u.username, u.created_at, u.updated_at
		FROM activities a
		JOIN users u ON a.user_id = u.id
		WHERE a.issue_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, issueID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanActivities(rows)
}

// List lists activities with filtering
func (r *ActivityRepository) List(ctx context.Context, filter *models.ActivityFilter) ([]*models.Activity, error) {
	query := `
		SELECT
			a.id, a.project_id, a.issue_id, a.user_id, a.action, a.entity_type,
			a.entity_id, a.field_name, a.old_value, a.new_value, a.ip_address,
			a.user_agent, a.metadata, a.created_at,
			u.id, u.email, u.username, u.created_at, u.updated_at
		FROM activities a
		JOIN users u ON a.user_id = u.id
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	// Build WHERE clause
	conditions := []string{}

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("a.project_id = $%d", argCount))
		args = append(args, *filter.ProjectID)
		argCount++
	}

	if filter.IssueID != nil {
		conditions = append(conditions, fmt.Sprintf("a.issue_id = $%d", argCount))
		args = append(args, *filter.IssueID)
		argCount++
	}

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("a.user_id = $%d", argCount))
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argCount))
		args = append(args, *filter.Action)
		argCount++
	}

	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("a.entity_type = $%d", argCount))
		args = append(args, *filter.EntityType)
		argCount++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY a.created_at DESC"

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanActivities(rows)
}

// scanActivities scans rows into activity models
func (r *ActivityRepository) scanActivities(rows *sql.Rows) ([]*models.Activity, error) {
	activities := make([]*models.Activity, 0)

	for rows.Next() {
		activity := &models.Activity{
			User: &models.User{},
		}

		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.IssueID,
			&activity.UserID,
			&activity.Action,
			&activity.EntityType,
			&activity.EntityID,
			&activity.FieldName,
			&activity.OldValue,
			&activity.NewValue,
			&activity.IPAddress,
			&activity.UserAgent,
			&activity.Metadata,
			&activity.CreatedAt,
			&activity.User.ID,
			&activity.User.Email,
			&activity.User.Username,
			&activity.User.CreatedAt,
			&activity.User.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}
