package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// CommentRepository handles comment data access
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	query := `
		INSERT INTO comments (issue_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, issue_id, user_id, content, created_at, updated_at
	`

	var created models.Comment
	err := r.db.QueryRowContext(ctx, query,
		comment.IssueID,
		comment.UserID,
		comment.Content,
	).Scan(
		&created.ID,
		&created.IssueID,
		&created.UserID,
		&created.Content,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id int) (*models.Comment, error) {
	query := `
		SELECT id, issue_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var comment models.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.IssueID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &comment, nil
}

// ListByIssueID retrieves all comments for an issue
func (r *CommentRepository) ListByIssueID(ctx context.Context, issueID int) ([]*models.Comment, error) {
	query := `
		SELECT id, issue_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE issue_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*models.Comment, 0)
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.IssueID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.ID)
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

// Delete deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`

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
