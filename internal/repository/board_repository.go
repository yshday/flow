package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// BoardRepository handles board column data access
type BoardRepository struct {
	db *sql.DB
}

// NewBoardRepository creates a new board repository
func NewBoardRepository(db *sql.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

// CreateDefaultColumns creates default columns for a project (Backlog, In Progress, Done)
func (r *BoardRepository) CreateDefaultColumns(ctx context.Context, projectID int) error {
	defaultColumns := []struct {
		name     string
		position int
	}{
		{"Backlog", 0},
		{"In Progress", 1},
		{"Done", 2},
	}

	for _, col := range defaultColumns {
		query := `
			INSERT INTO board_columns (project_id, name, position)
			VALUES ($1, $2, $3)
		`
		_, err := r.db.ExecContext(ctx, query, projectID, col.name, col.position)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListByProjectID retrieves all columns for a project
func (r *BoardRepository) ListByProjectID(ctx context.Context, projectID int) ([]*models.BoardColumn, error) {
	query := `
		SELECT id, project_id, name, position, created_at
		FROM board_columns
		WHERE project_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*models.BoardColumn, 0)
	for rows.Next() {
		var column models.BoardColumn
		err := rows.Scan(
			&column.ID,
			&column.ProjectID,
			&column.Name,
			&column.Position,
			&column.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		columns = append(columns, &column)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// Create creates a new board column
func (r *BoardRepository) Create(ctx context.Context, column *models.BoardColumn) (*models.BoardColumn, error) {
	query := `
		INSERT INTO board_columns (project_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, name, position, created_at
	`

	var created models.BoardColumn
	err := r.db.QueryRowContext(ctx, query,
		column.ProjectID,
		column.Name,
		column.Position,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.Name,
		&created.Position,
		&created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a board column by ID
func (r *BoardRepository) GetByID(ctx context.Context, id int) (*models.BoardColumn, error) {
	query := `
		SELECT id, project_id, name, position, created_at
		FROM board_columns
		WHERE id = $1
	`

	var column models.BoardColumn
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&column.ID,
		&column.ProjectID,
		&column.Name,
		&column.Position,
		&column.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &column, nil
}

// Update updates a board column
func (r *BoardRepository) Update(ctx context.Context, column *models.BoardColumn) error {
	query := `
		UPDATE board_columns
		SET name = $1, position = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, column.Name, column.Position, column.ID)
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

// Delete deletes a board column
func (r *BoardRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM board_columns WHERE id = $1`

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
