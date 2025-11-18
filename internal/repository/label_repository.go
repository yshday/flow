package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// LabelRepository handles label data access
type LabelRepository struct {
	db *sql.DB
}

// NewLabelRepository creates a new label repository
func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

// Create creates a new label
func (r *LabelRepository) Create(ctx context.Context, label *models.Label) (*models.Label, error) {
	query := `
		INSERT INTO labels (project_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, name, color, created_at
	`

	var created models.Label
	err := r.db.QueryRowContext(ctx, query,
		label.ProjectID,
		label.Name,
		label.Color,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.Name,
		&created.Color,
		&created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a label by ID
func (r *LabelRepository) GetByID(ctx context.Context, id int) (*models.Label, error) {
	query := `
		SELECT id, project_id, name, color, created_at
		FROM labels
		WHERE id = $1
	`

	var label models.Label
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&label.ID,
		&label.ProjectID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &label, nil
}

// ListByProjectID retrieves all labels for a project
func (r *LabelRepository) ListByProjectID(ctx context.Context, projectID int) ([]*models.Label, error) {
	query := `
		SELECT id, project_id, name, color, created_at
		FROM labels
		WHERE project_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := make([]*models.Label, 0)
	for rows.Next() {
		var label models.Label
		err := rows.Scan(
			&label.ID,
			&label.ProjectID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		labels = append(labels, &label)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}

// Update updates a label
func (r *LabelRepository) Update(ctx context.Context, label *models.Label) error {
	query := `
		UPDATE labels
		SET name = $1, color = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, label.Name, label.Color, label.ID)
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

// Delete deletes a label
func (r *LabelRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM labels WHERE id = $1`

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

// AddToIssue adds a label to an issue
func (r *LabelRepository) AddToIssue(ctx context.Context, issueID int, labelID int) error {
	query := `
		INSERT INTO issue_labels (issue_id, label_id)
		VALUES ($1, $2)
		ON CONFLICT (issue_id, label_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, issueID, labelID)
	return err
}

// RemoveFromIssue removes a label from an issue
func (r *LabelRepository) RemoveFromIssue(ctx context.Context, issueID int, labelID int) error {
	query := `
		DELETE FROM issue_labels
		WHERE issue_id = $1 AND label_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, issueID, labelID)
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

// ListByIssueID retrieves all labels for an issue
func (r *LabelRepository) ListByIssueID(ctx context.Context, issueID int) ([]*models.Label, error) {
	query := `
		SELECT l.id, l.project_id, l.name, l.color, l.created_at
		FROM labels l
		INNER JOIN issue_labels il ON l.id = il.label_id
		WHERE il.issue_id = $1
		ORDER BY l.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := make([]*models.Label, 0)
	for rows.Next() {
		var label models.Label
		err := rows.Scan(
			&label.ID,
			&label.ProjectID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		labels = append(labels, &label)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}
