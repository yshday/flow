package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// TasklistRepository handles tasklist item data access
type TasklistRepository struct {
	db *sql.DB
}

// NewTasklistRepository creates a new tasklist repository
func NewTasklistRepository(db *sql.DB) *TasklistRepository {
	return &TasklistRepository{db: db}
}

// Create creates a new tasklist item
func (r *TasklistRepository) Create(ctx context.Context, item *models.TasklistItem) (*models.TasklistItem, error) {
	// Get the next position if not specified
	if item.Position == 0 {
		var maxPos sql.NullInt64
		err := r.db.QueryRowContext(ctx,
			"SELECT MAX(position) FROM tasklist_items WHERE issue_id = $1",
			item.IssueID,
		).Scan(&maxPos)
		if err != nil {
			return nil, err
		}
		if maxPos.Valid {
			item.Position = int(maxPos.Int64) + 1
		}
	}

	query := `
		INSERT INTO tasklist_items (issue_id, content, is_completed, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
	`

	var created models.TasklistItem
	err := r.db.QueryRowContext(ctx, query,
		item.IssueID,
		item.Content,
		item.IsCompleted,
		item.Position,
	).Scan(
		&created.ID,
		&created.IssueID,
		&created.Content,
		&created.IsCompleted,
		&created.Position,
		&created.CompletedAt,
		&created.CompletedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a tasklist item by ID
func (r *TasklistRepository) GetByID(ctx context.Context, id int) (*models.TasklistItem, error) {
	query := `
		SELECT id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
		FROM tasklist_items
		WHERE id = $1
	`

	var item models.TasklistItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.IssueID,
		&item.Content,
		&item.IsCompleted,
		&item.Position,
		&item.CompletedAt,
		&item.CompletedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

// ListByIssueID retrieves all tasklist items for an issue
func (r *TasklistRepository) ListByIssueID(ctx context.Context, issueID int) ([]*models.TasklistItem, error) {
	query := `
		SELECT id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
		FROM tasklist_items
		WHERE issue_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.TasklistItem, 0)
	for rows.Next() {
		var item models.TasklistItem
		err := rows.Scan(
			&item.ID,
			&item.IssueID,
			&item.Content,
			&item.IsCompleted,
			&item.Position,
			&item.CompletedAt,
			&item.CompletedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

// Update updates a tasklist item
func (r *TasklistRepository) Update(ctx context.Context, item *models.TasklistItem) (*models.TasklistItem, error) {
	query := `
		UPDATE tasklist_items
		SET content = $1, is_completed = $2, position = $3, completed_at = $4, completed_by = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
	`

	var updated models.TasklistItem
	err := r.db.QueryRowContext(ctx, query,
		item.Content,
		item.IsCompleted,
		item.Position,
		item.CompletedAt,
		item.CompletedBy,
		item.ID,
	).Scan(
		&updated.ID,
		&updated.IssueID,
		&updated.Content,
		&updated.IsCompleted,
		&updated.Position,
		&updated.CompletedAt,
		&updated.CompletedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &updated, nil
}

// Delete removes a tasklist item
func (r *TasklistRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tasklist_items WHERE id = $1`

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

// Toggle toggles the completion status of a tasklist item
func (r *TasklistRepository) Toggle(ctx context.Context, id int, userID int) (*models.TasklistItem, error) {
	query := `
		UPDATE tasklist_items
		SET
			is_completed = NOT is_completed,
			completed_at = CASE WHEN NOT is_completed THEN NOW() ELSE NULL END,
			completed_by = CASE WHEN NOT is_completed THEN $1::INTEGER ELSE NULL END,
			updated_at = NOW()
		WHERE id = $2
		RETURNING id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
	`

	var updated models.TasklistItem
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(
		&updated.ID,
		&updated.IssueID,
		&updated.Content,
		&updated.IsCompleted,
		&updated.Position,
		&updated.CompletedAt,
		&updated.CompletedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &updated, nil
}

// Reorder updates positions of tasklist items
func (r *TasklistRepository) Reorder(ctx context.Context, issueID int, itemIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE tasklist_items SET position = $1, updated_at = NOW() WHERE id = $2 AND issue_id = $3`

	for i, itemID := range itemIDs {
		result, err := tx.ExecContext(ctx, query, i, itemID, issueID)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return pkgerrors.ErrNotFound
		}
	}

	return tx.Commit()
}

// GetProgress returns the completion progress of a tasklist
func (r *TasklistRepository) GetProgress(ctx context.Context, issueID int) (*models.TasklistProgress, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_completed = true) as completed
		FROM tasklist_items
		WHERE issue_id = $1
	`

	var total, completed int
	err := r.db.QueryRowContext(ctx, query, issueID).Scan(&total, &completed)
	if err != nil {
		return nil, err
	}

	progress := &models.TasklistProgress{
		Total:     total,
		Completed: completed,
		Pending:   total - completed,
		Percent:   0,
	}

	if total > 0 {
		progress.Percent = (completed * 100) / total
	}

	return progress, nil
}

// BulkCreate creates multiple tasklist items
func (r *TasklistRepository) BulkCreate(ctx context.Context, issueID int, items []models.CreateTasklistItemRequest) ([]*models.TasklistItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the current max position
	var maxPos sql.NullInt64
	err = tx.QueryRowContext(ctx,
		"SELECT MAX(position) FROM tasklist_items WHERE issue_id = $1",
		issueID,
	).Scan(&maxPos)
	if err != nil {
		return nil, err
	}

	startPos := 0
	if maxPos.Valid {
		startPos = int(maxPos.Int64) + 1
	}

	query := `
		INSERT INTO tasklist_items (issue_id, content, is_completed, position)
		VALUES ($1, $2, false, $3)
		RETURNING id, issue_id, content, is_completed, position, completed_at, completed_by, created_at, updated_at
	`

	created := make([]*models.TasklistItem, 0, len(items))
	for i, item := range items {
		var newItem models.TasklistItem
		err := tx.QueryRowContext(ctx, query, issueID, item.Content, startPos+i).Scan(
			&newItem.ID,
			&newItem.IssueID,
			&newItem.Content,
			&newItem.IsCompleted,
			&newItem.Position,
			&newItem.CompletedAt,
			&newItem.CompletedBy,
			&newItem.CreatedAt,
			&newItem.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		created = append(created, &newItem)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return created, nil
}
