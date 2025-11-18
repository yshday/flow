package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// MilestoneRepository handles milestone data access
type MilestoneRepository struct {
	db *sql.DB
}

// NewMilestoneRepository creates a new milestone repository
func NewMilestoneRepository(db *sql.DB) *MilestoneRepository {
	return &MilestoneRepository{db: db}
}

// Create creates a new milestone
func (r *MilestoneRepository) Create(ctx context.Context, milestone *models.Milestone) (*models.Milestone, error) {
	query := `
		INSERT INTO milestones (project_id, title, description, due_date, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		milestone.ProjectID,
		milestone.Title,
		milestone.Description,
		milestone.DueDate,
		milestone.Status,
	).Scan(&milestone.ID, &milestone.CreatedAt, &milestone.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return milestone, nil
}

// GetByID retrieves a milestone by ID
func (r *MilestoneRepository) GetByID(ctx context.Context, id int) (*models.Milestone, error) {
	query := `
		SELECT id, project_id, title, description, due_date, status, created_at, updated_at
		FROM milestones
		WHERE id = $1
	`

	milestone := &models.Milestone{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&milestone.ID,
		&milestone.ProjectID,
		&milestone.Title,
		&milestone.Description,
		&milestone.DueDate,
		&milestone.Status,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return milestone, nil
}

// ListByProjectID lists all milestones for a project
func (r *MilestoneRepository) ListByProjectID(ctx context.Context, projectID int) ([]*models.Milestone, error) {
	query := `
		SELECT id, project_id, title, description, due_date, status, created_at, updated_at
		FROM milestones
		WHERE project_id = $1
		ORDER BY due_date ASC NULLS LAST, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	milestones := make([]*models.Milestone, 0)
	for rows.Next() {
		milestone := &models.Milestone{}
		err := rows.Scan(
			&milestone.ID,
			&milestone.ProjectID,
			&milestone.Title,
			&milestone.Description,
			&milestone.DueDate,
			&milestone.Status,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		milestones = append(milestones, milestone)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return milestones, nil
}

// Update updates a milestone
func (r *MilestoneRepository) Update(ctx context.Context, milestone *models.Milestone) (*models.Milestone, error) {
	query := `
		UPDATE milestones
		SET title = $1, description = $2, due_date = $3, status = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		milestone.Title,
		milestone.Description,
		milestone.DueDate,
		milestone.Status,
		milestone.ID,
	).Scan(&milestone.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return milestone, nil
}

// Delete deletes a milestone
func (r *MilestoneRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM milestones WHERE id = $1`

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

// GetWithProgress retrieves a milestone with progress calculation
func (r *MilestoneRepository) GetWithProgress(ctx context.Context, id int) (*models.Milestone, error) {
	query := `
		SELECT
			m.id, m.project_id, m.title, m.description, m.due_date, m.status, m.created_at, m.updated_at,
			COUNT(i.id) AS total_issues,
			COUNT(CASE WHEN i.status = 'closed' THEN 1 END) AS closed_issues
		FROM milestones m
		LEFT JOIN issues i ON m.id = i.milestone_id AND i.deleted_at IS NULL
		WHERE m.id = $1
		GROUP BY m.id, m.project_id, m.title, m.description, m.due_date, m.status, m.created_at, m.updated_at
	`

	milestone := &models.Milestone{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&milestone.ID,
		&milestone.ProjectID,
		&milestone.Title,
		&milestone.Description,
		&milestone.DueDate,
		&milestone.Status,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
		&milestone.TotalIssues,
		&milestone.ClosedIssues,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	// Calculate progress percentage
	if milestone.TotalIssues > 0 {
		milestone.Progress = (milestone.ClosedIssues * 100) / milestone.TotalIssues
	} else {
		milestone.Progress = 0
	}

	return milestone, nil
}
