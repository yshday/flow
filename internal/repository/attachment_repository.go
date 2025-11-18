package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/pkg/errors"
)

// AttachmentRepository handles attachment database operations
type AttachmentRepository struct {
	db *sql.DB
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create creates a new attachment
func (r *AttachmentRepository) Create(ctx context.Context, attachment *models.Attachment) (*models.Attachment, error) {
	query := `
		INSERT INTO attachments (issue_id, user_id, storage_key, original_filename, file_size, content_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		attachment.IssueID,
		attachment.UserID,
		attachment.StorageKey,
		attachment.OriginalFilename,
		attachment.FileSize,
		attachment.ContentType,
	).Scan(&attachment.ID, &attachment.CreatedAt, &attachment.UpdatedAt)

	if err != nil {
		return nil, errors.NewInternalError("failed to create attachment in database", err)
	}

	return attachment, nil
}

// CreateWithTx creates a new attachment within a transaction
func (r *AttachmentRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, attachment *models.Attachment) (*models.Attachment, error) {
	query := `
		INSERT INTO attachments (issue_id, user_id, storage_key, original_filename, file_size, content_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := tx.QueryRowContext(
		ctx,
		query,
		attachment.IssueID,
		attachment.UserID,
		attachment.StorageKey,
		attachment.OriginalFilename,
		attachment.FileSize,
		attachment.ContentType,
	).Scan(&attachment.ID, &attachment.CreatedAt, &attachment.UpdatedAt)

	if err != nil {
		return nil, errors.NewInternalError("failed to create attachment in database", err)
	}

	return attachment, nil
}

// GetByID retrieves an attachment by ID
func (r *AttachmentRepository) GetByID(ctx context.Context, id int) (*models.Attachment, error) {
	query := `
		SELECT id, issue_id, user_id, storage_key, original_filename,
		       file_size, content_type, created_at, updated_at, deleted_at
		FROM attachments
		WHERE id = $1 AND deleted_at IS NULL
	`

	attachment := &models.Attachment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&attachment.ID,
		&attachment.IssueID,
		&attachment.UserID,
		&attachment.StorageKey,
		&attachment.OriginalFilename,
		&attachment.FileSize,
		&attachment.ContentType,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError(fmt.Sprintf("attachment with ID %d not found", id))
	}
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve attachment from database", err)
	}

	return attachment, nil
}

// ListByIssueID retrieves all attachments for an issue
func (r *AttachmentRepository) ListByIssueID(ctx context.Context, issueID int) ([]*models.Attachment, error) {
	query := `
		SELECT id, issue_id, user_id, storage_key, original_filename,
		       file_size, content_type, created_at, updated_at, deleted_at
		FROM attachments
		WHERE issue_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list attachments for issue %d", issueID), err)
	}
	defer rows.Close()

	attachments := make([]*models.Attachment, 0)
	for rows.Next() {
		attachment := &models.Attachment{}
		err := rows.Scan(
			&attachment.ID,
			&attachment.IssueID,
			&attachment.UserID,
			&attachment.StorageKey,
			&attachment.OriginalFilename,
			&attachment.FileSize,
			&attachment.ContentType,
			&attachment.CreatedAt,
			&attachment.UpdatedAt,
			&attachment.DeletedAt,
		)
		if err != nil {
			return nil, errors.NewInternalError("failed to scan attachment row", err)
		}
		attachments = append(attachments, attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternalError("error iterating attachment rows", err)
	}

	return attachments, nil
}

// Delete soft deletes an attachment
func (r *AttachmentRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE attachments
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to delete attachment with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("attachment with ID %d not found or already deleted", id))
	}

	return nil
}

// BeginTx starts a new database transaction
func (r *AttachmentRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.NewInternalError("failed to begin transaction", err)
	}
	return tx, nil
}
