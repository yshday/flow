package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
)

// MentionRepository handles mention data access
type MentionRepository struct {
	db *sql.DB
}

// NewMentionRepository creates a new mention repository
func NewMentionRepository(db *sql.DB) *MentionRepository {
	return &MentionRepository{db: db}
}

// Create creates a new mention
func (r *MentionRepository) Create(ctx context.Context, mention *models.Mention) error {
	query := `
		INSERT INTO mentions (user_id, mentioned_by_user_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, entity_type, entity_id) DO NOTHING
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		mention.UserID,
		mention.MentionedByUserID,
		mention.EntityType,
		mention.EntityID,
		mention.CreatedAt,
	).Scan(&mention.ID)

	if err == sql.ErrNoRows {
		// Mention already exists, not an error
		return nil
	}

	return err
}

// CreateBatch creates multiple mentions in a single transaction
func (r *MentionRepository) CreateBatch(ctx context.Context, mentions []*models.Mention) error {
	if len(mentions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO mentions (user_id, mentioned_by_user_id, entity_type, entity_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, entity_type, entity_id) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, mention := range mentions {
		_, err = stmt.ExecContext(ctx,
			mention.UserID,
			mention.MentionedByUserID,
			mention.EntityType,
			mention.EntityID,
			mention.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByEntity retrieves all mentions for a specific entity
func (r *MentionRepository) GetByEntity(ctx context.Context, entityType string, entityID int) ([]*models.Mention, error) {
	query := `
		SELECT m.id, m.user_id, m.mentioned_by_user_id, m.entity_type, m.entity_id, m.created_at,
		       u.id, u.username, u.email, u.avatar_url
		FROM mentions m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.entity_type = $1 AND m.entity_id = $2
		ORDER BY m.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mentions := make([]*models.Mention, 0)
	for rows.Next() {
		mention := &models.Mention{User: &models.User{}}
		err := rows.Scan(
			&mention.ID,
			&mention.UserID,
			&mention.MentionedByUserID,
			&mention.EntityType,
			&mention.EntityID,
			&mention.CreatedAt,
			&mention.User.ID,
			&mention.User.Username,
			&mention.User.Email,
			&mention.User.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		mentions = append(mentions, mention)
	}

	return mentions, rows.Err()
}

// GetByUser retrieves all mentions for a specific user
func (r *MentionRepository) GetByUser(ctx context.Context, userID int, limit, offset int) ([]*models.Mention, error) {
	query := `
		SELECT m.id, m.user_id, m.mentioned_by_user_id, m.entity_type, m.entity_id, m.created_at,
		       u.id, u.username, u.email, u.avatar_url
		FROM mentions m
		LEFT JOIN users u ON m.mentioned_by_user_id = u.id
		WHERE m.user_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mentions := make([]*models.Mention, 0)
	for rows.Next() {
		mention := &models.Mention{MentionedBy: &models.User{}}
		err := rows.Scan(
			&mention.ID,
			&mention.UserID,
			&mention.MentionedByUserID,
			&mention.EntityType,
			&mention.EntityID,
			&mention.CreatedAt,
			&mention.MentionedBy.ID,
			&mention.MentionedBy.Username,
			&mention.MentionedBy.Email,
			&mention.MentionedBy.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		mentions = append(mentions, mention)
	}

	return mentions, rows.Err()
}

// DeleteByEntity deletes all mentions for a specific entity
func (r *MentionRepository) DeleteByEntity(ctx context.Context, entityType string, entityID int) error {
	query := `DELETE FROM mentions WHERE entity_type = $1 AND entity_id = $2`
	_, err := r.db.ExecContext(ctx, query, entityType, entityID)
	return err
}
