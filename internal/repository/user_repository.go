package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, username, password_hash, name, avatar_url, external_id, external_provider)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, username, name, password_hash, avatar_url, external_id, external_provider, created_at, updated_at
	`

	var createdUser models.User
	err := r.db.QueryRowContext(ctx, query,
		user.Email, user.Username, user.PasswordHash,
		user.Name, user.AvatarURL, user.ExternalID, user.ExternalProvider,
	).Scan(
		&createdUser.ID,
		&createdUser.Email,
		&createdUser.Username,
		&createdUser.Name,
		&createdUser.PasswordHash,
		&createdUser.AvatarURL,
		&createdUser.ExternalID,
		&createdUser.ExternalProvider,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// 23505 is the PostgreSQL error code for unique violation
			if pqErr.Code == "23505" {
				return nil, pkgerrors.ErrConflict
			}
		}
		return nil, err
	}

	return &createdUser, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, name, password_hash, avatar_url, external_id, external_provider, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Name,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.ExternalID,
		&user.ExternalProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, email, username, name, password_hash, avatar_url, external_id, external_provider, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Name,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.ExternalID,
		&user.ExternalProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByExternalID retrieves a user by external ID and provider
func (r *UserRepository) GetByExternalID(ctx context.Context, externalID, provider string) (*models.User, error) {
	query := `
		SELECT id, email, username, name, password_hash, avatar_url, external_id, external_provider, created_at, updated_at
		FROM users
		WHERE external_id = $1 AND external_provider = $2
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, externalID, provider).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Name,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.ExternalID,
		&user.ExternalProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.Username, user.AvatarURL, user.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return pkgerrors.ErrConflict
			}
		}
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

// Search searches for users by email or username
func (r *UserRepository) Search(ctx context.Context, query string, limit int) ([]*models.User, error) {
	searchQuery := `
		SELECT id, email, username, avatar_url, created_at, updated_at
		FROM users
		WHERE email ILIKE $1 OR username ILIKE $1
		ORDER BY username
		LIMIT $2
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
