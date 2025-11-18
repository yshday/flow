package repository

import (
	"database/sql"
	"fmt"

	"github.com/yourusername/issue-tracker/internal/models"
)

// ReactionRepository handles reactions data access
type ReactionRepository struct {
	db *sql.DB
}

// NewReactionRepository creates a new ReactionRepository
func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

// Create adds a new reaction
func (r *ReactionRepository) Create(reaction *models.Reaction) error {
	query := `
		INSERT INTO reactions (user_id, entity_type, entity_id, emoji, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		reaction.UserID,
		reaction.EntityType,
		reaction.EntityID,
		reaction.Emoji,
		reaction.CreatedAt,
	).Scan(&reaction.ID, &reaction.CreatedAt)
}

// Delete removes a reaction
func (r *ReactionRepository) Delete(userID int, entityType string, entityID int, emoji string) error {
	query := `
		DELETE FROM reactions
		WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3 AND emoji = $4
	`
	result, err := r.db.Exec(query, userID, entityType, entityID, emoji)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reaction not found")
	}

	return nil
}

// GetByEntity retrieves all reactions for a specific entity
func (r *ReactionRepository) GetByEntity(entityType string, entityID int) ([]*models.Reaction, error) {
	query := `
		SELECT r.id, r.user_id, r.entity_type, r.entity_id, r.emoji, r.created_at,
		       u.id, u.username, u.email
		FROM reactions r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.entity_type = $1 AND r.entity_id = $2
		ORDER BY r.created_at ASC
	`

	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make([]*models.Reaction, 0)
	for rows.Next() {
		reaction := &models.Reaction{User: &models.User{}}
		var email sql.NullString

		err := rows.Scan(
			&reaction.ID,
			&reaction.UserID,
			&reaction.EntityType,
			&reaction.EntityID,
			&reaction.Emoji,
			&reaction.CreatedAt,
			&reaction.User.ID,
			&reaction.User.Username,
			&email,
		)
		if err != nil {
			return nil, err
		}

		if email.Valid {
			reaction.User.Email = email.String
		}

		reactions = append(reactions, reaction)
	}

	return reactions, rows.Err()
}

// GetSummary returns aggregated reaction counts for an entity
func (r *ReactionRepository) GetSummary(entityType string, entityID int) (*models.ReactionSummary, error) {
	query := `
		SELECT emoji, user_id
		FROM reactions
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY emoji, user_id
	`

	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := &models.ReactionSummary{
		EntityType:    entityType,
		EntityID:      entityID,
		Reactions:     make(map[string]int),
		UserReactions: make(map[string][]int),
	}

	for rows.Next() {
		var emoji string
		var userID int

		if err := rows.Scan(&emoji, &userID); err != nil {
			return nil, err
		}

		summary.Reactions[emoji]++
		summary.UserReactions[emoji] = append(summary.UserReactions[emoji], userID)
	}

	return summary, rows.Err()
}

// GetByUser retrieves a user's reaction for a specific entity and emoji
func (r *ReactionRepository) GetByUser(userID int, entityType string, entityID int, emoji string) (*models.Reaction, error) {
	query := `
		SELECT id, user_id, entity_type, entity_id, emoji, created_at
		FROM reactions
		WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3 AND emoji = $4
	`

	reaction := &models.Reaction{}
	err := r.db.QueryRow(query, userID, entityType, entityID, emoji).Scan(
		&reaction.ID,
		&reaction.UserID,
		&reaction.EntityType,
		&reaction.EntityID,
		&reaction.Emoji,
		&reaction.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return reaction, nil
}

// DeleteByEntity removes all reactions for a specific entity
func (r *ReactionRepository) DeleteByEntity(entityType string, entityID int) error {
	query := `
		DELETE FROM reactions
		WHERE entity_type = $1 AND entity_id = $2
	`
	_, err := r.db.Exec(query, entityType, entityID)
	return err
}
