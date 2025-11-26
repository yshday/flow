package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// TemplateRepository handles template data access
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// ==================== Project Templates ====================

// ListProjectTemplates returns all project templates
func (r *TemplateRepository) ListProjectTemplates(ctx context.Context) ([]*models.ProjectTemplate, error) {
	query := `
		SELECT id, name, description, is_system, created_by, config, created_at, updated_at
		FROM project_templates
		ORDER BY is_system DESC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.ProjectTemplate, 0)
	for rows.Next() {
		var t models.ProjectTemplate
		var description sql.NullString
		var createdBy sql.NullInt64
		var configJSON []byte

		err := rows.Scan(
			&t.ID,
			&t.Name,
			&description,
			&t.IsSystem,
			&createdBy,
			&configJSON,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			t.CreatedBy = &createdByInt
		}

		if err := json.Unmarshal(configJSON, &t.Config); err != nil {
			return nil, err
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// GetProjectTemplate returns a project template by ID
func (r *TemplateRepository) GetProjectTemplate(ctx context.Context, id int) (*models.ProjectTemplate, error) {
	query := `
		SELECT id, name, description, is_system, created_by, config, created_at, updated_at
		FROM project_templates
		WHERE id = $1
	`

	var t models.ProjectTemplate
	var description sql.NullString
	var createdBy sql.NullInt64
	var configJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Name,
		&description,
		&t.IsSystem,
		&createdBy,
		&configJSON,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	if description.Valid {
		t.Description = description.String
	}
	if createdBy.Valid {
		createdByInt := int(createdBy.Int64)
		t.CreatedBy = &createdByInt
	}

	if err := json.Unmarshal(configJSON, &t.Config); err != nil {
		return nil, err
	}

	return &t, nil
}

// CreateProjectTemplate creates a new project template
func (r *TemplateRepository) CreateProjectTemplate(ctx context.Context, t *models.ProjectTemplate) (*models.ProjectTemplate, error) {
	configJSON, err := json.Marshal(t.Config)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO project_templates (name, description, is_system, created_by, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		t.Name,
		t.Description,
		t.IsSystem,
		t.CreatedBy,
		configJSON,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// ==================== Issue Templates ====================

// ListIssueTemplates returns all issue templates for a project
func (r *TemplateRepository) ListIssueTemplates(ctx context.Context, projectID int) ([]*models.IssueTemplate, error) {
	query := `
		SELECT id, project_id, name, description, content, default_priority,
		       default_labels, position, is_active, created_by, created_at, updated_at
		FROM issue_templates
		WHERE project_id = $1
		ORDER BY position ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.IssueTemplate, 0)
	for rows.Next() {
		var t models.IssueTemplate
		var description sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&description,
			&t.Content,
			&t.DefaultPriority,
			pq.Array(&t.DefaultLabels),
			&t.Position,
			&t.IsActive,
			&createdBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			t.CreatedBy = &createdByInt
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// ListActiveIssueTemplates returns only active issue templates for a project
func (r *TemplateRepository) ListActiveIssueTemplates(ctx context.Context, projectID int) ([]*models.IssueTemplate, error) {
	query := `
		SELECT id, project_id, name, description, content, default_priority,
		       default_labels, position, is_active, created_by, created_at, updated_at
		FROM issue_templates
		WHERE project_id = $1 AND is_active = true
		ORDER BY position ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.IssueTemplate, 0)
	for rows.Next() {
		var t models.IssueTemplate
		var description sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&description,
			&t.Content,
			&t.DefaultPriority,
			pq.Array(&t.DefaultLabels),
			&t.Position,
			&t.IsActive,
			&createdBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if createdBy.Valid {
			createdByInt := int(createdBy.Int64)
			t.CreatedBy = &createdByInt
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// GetIssueTemplate returns an issue template by ID
func (r *TemplateRepository) GetIssueTemplate(ctx context.Context, id int) (*models.IssueTemplate, error) {
	query := `
		SELECT id, project_id, name, description, content, default_priority,
		       default_labels, position, is_active, created_by, created_at, updated_at
		FROM issue_templates
		WHERE id = $1
	`

	var t models.IssueTemplate
	var description sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.ProjectID,
		&t.Name,
		&description,
		&t.Content,
		&t.DefaultPriority,
		pq.Array(&t.DefaultLabels),
		&t.Position,
		&t.IsActive,
		&createdBy,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	if description.Valid {
		t.Description = description.String
	}
	if createdBy.Valid {
		createdByInt := int(createdBy.Int64)
		t.CreatedBy = &createdByInt
	}

	return &t, nil
}

// CreateIssueTemplate creates a new issue template
func (r *TemplateRepository) CreateIssueTemplate(ctx context.Context, t *models.IssueTemplate) (*models.IssueTemplate, error) {
	query := `
		INSERT INTO issue_templates (project_id, name, description, content, default_priority, default_labels, position, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		t.ProjectID,
		t.Name,
		t.Description,
		t.Content,
		t.DefaultPriority,
		pq.Array(t.DefaultLabels),
		t.Position,
		t.IsActive,
		t.CreatedBy,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// UpdateIssueTemplate updates an issue template
func (r *TemplateRepository) UpdateIssueTemplate(ctx context.Context, t *models.IssueTemplate) error {
	query := `
		UPDATE issue_templates
		SET name = $1, description = $2, content = $3, default_priority = $4,
		    default_labels = $5, position = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		t.Name,
		t.Description,
		t.Content,
		t.DefaultPriority,
		pq.Array(t.DefaultLabels),
		t.Position,
		t.IsActive,
		t.ID,
	)
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

// DeleteIssueTemplate deletes an issue template
func (r *TemplateRepository) DeleteIssueTemplate(ctx context.Context, id int) error {
	query := `DELETE FROM issue_templates WHERE id = $1`

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
