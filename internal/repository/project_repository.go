package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ProjectRepository handles project data access
type ProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) (*models.Project, error) {
	query := `
		INSERT INTO projects (name, key, description, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, key, description, owner_id, created_at, updated_at
	`

	var created models.Project
	err := r.db.QueryRowContext(ctx, query,
		project.Name,
		project.Key,
		project.Description,
		project.OwnerID,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Key,
		&created.Description,
		&created.OwnerID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique violation
				return nil, pkgerrors.ErrConflict
			}
		}
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id int) (*models.Project, error) {
	query := `
		SELECT id, name, key, description, owner_id, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var project models.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Key,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &project, nil
}

// GetByKey retrieves a project by key
func (r *ProjectRepository) GetByKey(ctx context.Context, key string) (*models.Project, error) {
	query := `
		SELECT id, name, key, description, owner_id, created_at, updated_at
		FROM projects
		WHERE key = $1
	`

	var project models.Project
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&project.ID,
		&project.Name,
		&project.Key,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &project, nil
}

// ListByUserID retrieves all projects owned by a user
func (r *ProjectRepository) ListByUserID(ctx context.Context, userID int) ([]*models.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.key, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.owner_id = $1 OR pm.user_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]*models.Project, 0)
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Key,
			&project.Description,
			&project.OwnerID,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		project.Description,
		project.ID,
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

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM projects WHERE id = $1`

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
