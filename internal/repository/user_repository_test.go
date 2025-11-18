package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func cleanupUsers(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM users WHERE email LIKE 'test%@example.com'")
	if err != nil {
		t.Fatalf("Failed to cleanup users: %v", err)
	}
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("should create user successfully", func(t *testing.T) {
		user := &models.User{
			Email:        "test1@example.com",
			Username:     "testuser1",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		}

		createdUser, err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if createdUser.ID == 0 {
			t.Error("Expected ID to be set")
		}

		if createdUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, createdUser.Email)
		}

		if createdUser.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		user1 := &models.User{
			Email:        "test2@example.com",
			Username:     "testuser2",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		}

		_, err := repo.Create(ctx, user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create another user with same email
		user2 := &models.User{
			Email:        "test2@example.com",
			Username:     "testuser3",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		}

		_, err = repo.Create(ctx, user2)
		if err != pkgerrors.ErrConflict {
			t.Errorf("Expected ErrConflict, got %v", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("should get user by email", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			Email:        "test3@example.com",
			Username:     "testuser3",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		}

		createdUser, err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Get the user by email
		foundUser, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected ID %d, got %d", createdUser.ID, foundUser.ID)
		}

		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}
	})

	t.Run("should return ErrNotFound for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupUsers(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("should get user by ID", func(t *testing.T) {
		user := &models.User{
			Email:        "test4@example.com",
			Username:     "testuser4",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		}

		createdUser, err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		foundUser, err := repo.GetByID(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected ID %d, got %d", createdUser.ID, foundUser.ID)
		}
	})

	t.Run("should return ErrNotFound for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999999)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}
