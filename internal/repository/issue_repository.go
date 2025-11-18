package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// IssueRepository handles issue data access
type IssueRepository struct {
	db *sql.DB
}

// NewIssueRepository creates a new issue repository
func NewIssueRepository(db *sql.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

// Create creates a new issue with auto-generated issue number
func (r *IssueRepository) Create(ctx context.Context, issue *models.Issue) (*models.Issue, error) {
	query := `
		INSERT INTO issues (
			project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id
		)
		VALUES ($1, get_next_issue_number($1), $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id,
			version, created_at, updated_at, deleted_at
	`

	var created models.Issue
	err := r.db.QueryRowContext(ctx, query,
		issue.ProjectID,
		issue.Title,
		issue.Description,
		issue.Status,
		issue.ColumnID,
		issue.ColumnPosition,
		issue.Priority,
		issue.AssigneeID,
		issue.ReporterID,
		issue.MilestoneID,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.IssueNumber,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.ColumnID,
		&created.ColumnPosition,
		&created.Priority,
		&created.AssigneeID,
		&created.ReporterID,
		&created.MilestoneID,
		&created.Version,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves an issue by ID (excluding soft-deleted)
func (r *IssueRepository) GetByID(ctx context.Context, id int) (*models.Issue, error) {
	query := `
		SELECT id, project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id,
			version, created_at, updated_at, deleted_at
		FROM issues
		WHERE id = $1 AND deleted_at IS NULL
	`

	var issue models.Issue
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.ColumnID,
		&issue.ColumnPosition,
		&issue.Priority,
		&issue.AssigneeID,
		&issue.ReporterID,
		&issue.MilestoneID,
		&issue.Version,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &issue, nil
}

// GetByProjectAndNumber retrieves an issue by project ID and issue number
func (r *IssueRepository) GetByProjectAndNumber(ctx context.Context, projectID int, issueNumber int) (*models.Issue, error) {
	query := `
		SELECT id, project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id,
			version, created_at, updated_at, deleted_at
		FROM issues
		WHERE project_id = $1 AND issue_number = $2 AND deleted_at IS NULL
	`

	var issue models.Issue
	err := r.db.QueryRowContext(ctx, query, projectID, issueNumber).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.ColumnID,
		&issue.ColumnPosition,
		&issue.Priority,
		&issue.AssigneeID,
		&issue.ReporterID,
		&issue.MilestoneID,
		&issue.Version,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &issue, nil
}

// List retrieves issues with filtering and search
func (r *IssueRepository) List(ctx context.Context, filter *models.IssueFilter) ([]*models.Issue, error) {
	query := `
		SELECT id, project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id,
			version, created_at, updated_at, deleted_at
		FROM issues
		WHERE project_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{filter.ProjectID}
	argCount := 1

	// Add filters
	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	if filter.Priority != nil {
		argCount++
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filter.Priority)
	}

	if filter.AssigneeID != nil {
		argCount++
		query += fmt.Sprintf(" AND assignee_id = $%d", argCount)
		args = append(args, *filter.AssigneeID)
	}

	if filter.ReporterID != nil {
		argCount++
		query += fmt.Sprintf(" AND reporter_id = $%d", argCount)
		args = append(args, *filter.ReporterID)
	}

	if filter.MilestoneID != nil {
		argCount++
		query += fmt.Sprintf(" AND milestone_id = $%d", argCount)
		args = append(args, *filter.MilestoneID)
	}

	// Search by title or description
	if filter.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
	}

	// Label filtering (if provided)
	if len(filter.LabelIDs) > 0 {
		placeholders := make([]string, len(filter.LabelIDs))
		for i, labelID := range filter.LabelIDs {
			argCount++
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, labelID)
		}
		query += fmt.Sprintf(` AND id IN (
			SELECT issue_id FROM issue_labels WHERE label_id IN (%s)
		)`, strings.Join(placeholders, ","))
	}

	// Order by issue number descending (newest first)
	query += " ORDER BY issue_number DESC"

	// Pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*models.Issue, 0)
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(
			&issue.ID,
			&issue.ProjectID,
			&issue.IssueNumber,
			&issue.Title,
			&issue.Description,
			&issue.Status,
			&issue.ColumnID,
			&issue.ColumnPosition,
			&issue.Priority,
			&issue.AssigneeID,
			&issue.ReporterID,
			&issue.MilestoneID,
			&issue.Version,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, &issue)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

// Update updates an issue with optimistic locking
func (r *IssueRepository) Update(ctx context.Context, issue *models.Issue) error {
	query := `
		UPDATE issues
		SET title = $1, description = $2, status = $3, priority = $4,
			assignee_id = $5, milestone_id = $6, column_id = $7, column_position = $8,
			version = version + 1, updated_at = NOW()
		WHERE id = $9 AND version = $10 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		issue.Title,
		issue.Description,
		issue.Status,
		issue.Priority,
		issue.AssigneeID,
		issue.MilestoneID,
		issue.ColumnID,
		issue.ColumnPosition,
		issue.ID,
		issue.Version,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Check if issue exists
		var exists bool
		err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM issues WHERE id = $1 AND deleted_at IS NULL)", issue.ID).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return pkgerrors.ErrNotFound
		}
		// Issue exists but version mismatch
		return pkgerrors.ErrConflict
	}

	return nil
}

// Delete soft deletes an issue
func (r *IssueRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE issues
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// Search searches for issues by text in title and description
func (r *IssueRepository) Search(ctx context.Context, projectID int, query string, limit int, offset int) ([]*models.Issue, error) {
	searchQuery := `
		SELECT id, project_id, issue_number, title, description, status,
			column_id, column_position, priority, assignee_id, reporter_id, milestone_id,
			version, created_at, updated_at, deleted_at
		FROM issues
		WHERE project_id = $1
			AND deleted_at IS NULL
			AND (title ILIKE $2 OR description ILIKE $2)
		ORDER BY issue_number DESC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + query + "%"

	rows, err := r.db.QueryContext(ctx, searchQuery, projectID, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*models.Issue, 0)
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(
			&issue.ID,
			&issue.ProjectID,
			&issue.IssueNumber,
			&issue.Title,
			&issue.Description,
			&issue.Status,
			&issue.ColumnID,
			&issue.ColumnPosition,
			&issue.Priority,
			&issue.AssigneeID,
			&issue.ReporterID,
			&issue.MilestoneID,
			&issue.Version,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		issues = append(issues, &issue)
	}

	return issues, rows.Err()
}
