package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yourusername/issue-tracker/internal/models"
)

// SearchRepository handles search operations
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new search repository
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SearchIssues searches for issues based on criteria
func (r *SearchRepository) SearchIssues(ctx context.Context, req *models.IssueSearchRequest) ([]*models.IssueSearchResult, int, error) {
	// Build the WHERE clause dynamically
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Always filter out deleted issues
	whereClauses = append(whereClauses, "i.deleted_at IS NULL")

	// Text search in title and description
	if req.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(i.title ILIKE $%d OR i.description ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+req.Query+"%")
		argCount++
	}

	// Filter by project(s)
	// ProjectIDs takes precedence over ProjectID for permission filtering
	if len(req.ProjectIDs) > 0 {
		placeholders := make([]string, len(req.ProjectIDs))
		for i, projectID := range req.ProjectIDs {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, projectID)
			argCount++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("i.project_id IN (%s)", strings.Join(placeholders, ", ")))
	} else if req.ProjectID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("i.project_id = $%d", argCount))
		args = append(args, *req.ProjectID)
		argCount++
	}

	// Filter by status
	if len(req.Status) > 0 {
		placeholders := make([]string, len(req.Status))
		for i, status := range req.Status {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, status)
			argCount++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("i.status IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Filter by priority
	if len(req.Priority) > 0 {
		placeholders := make([]string, len(req.Priority))
		for i, priority := range req.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, priority)
			argCount++
		}
		whereClauses = append(whereClauses, fmt.Sprintf("i.priority IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Filter by assignee
	if req.AssigneeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("i.assignee_id = $%d", argCount))
		args = append(args, *req.AssigneeID)
		argCount++
	}

	// Filter by reporter
	if req.ReporterID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("i.reporter_id = $%d", argCount))
		args = append(args, *req.ReporterID)
		argCount++
	}

	// Filter by labels (issues that have ALL specified labels)
	if len(req.LabelIDs) > 0 {
		placeholders := make([]string, len(req.LabelIDs))
		for i, labelID := range req.LabelIDs {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, labelID)
			argCount++
		}
		whereClauses = append(whereClauses, fmt.Sprintf(`
			i.id IN (
				SELECT issue_id
				FROM issue_labels
				WHERE label_id IN (%s)
				GROUP BY issue_id
				HAVING COUNT(DISTINCT label_id) = %d
			)
		`, strings.Join(placeholders, ", "), len(req.LabelIDs)))
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM issues i
		WHERE %s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Set default limit and offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			i.id,
			i.project_id,
			p.key as project_key,
			i.issue_number,
			i.title,
			COALESCE(i.description, '') as description,
			i.status,
			i.priority,
			i.assignee_id,
			i.reporter_id,
			i.created_at,
			i.updated_at
		FROM issues i
		JOIN projects p ON i.project_id = p.id
		WHERE %s
		ORDER BY i.updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]*models.IssueSearchResult, 0)
	for rows.Next() {
		result := &models.IssueSearchResult{}
		err := rows.Scan(
			&result.ID,
			&result.ProjectID,
			&result.ProjectKey,
			&result.IssueNumber,
			&result.Title,
			&result.Description,
			&result.Status,
			&result.Priority,
			&result.AssigneeID,
			&result.ReporterID,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// SearchProjects searches for projects based on criteria
func (r *SearchRepository) SearchProjects(ctx context.Context, req *models.ProjectSearchRequest) ([]*models.ProjectSearchResult, int, error) {
	// Build the WHERE clause
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Projects table doesn't have deleted_at column (uses regular DELETE)
	// No need to filter for deleted projects

	// Text search in name, description, and key
	if req.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR key ILIKE $%d)", argCount, argCount, argCount))
		args = append(args, "%"+req.Query+"%")
		argCount++
	}

	// Build WHERE clause (if any)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM projects
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Set default limit and offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			key,
			COALESCE(description, '') as description,
			owner_id,
			created_at
		FROM projects
		%s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]*models.ProjectSearchResult, 0)
	for rows.Next() {
		result := &models.ProjectSearchResult{}
		err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Key,
			&result.Description,
			&result.OwnerID,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
