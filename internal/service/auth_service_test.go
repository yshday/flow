package service

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/issue-tracker/internal/auth"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupAuthService(t *testing.T) (*AuthService, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	jwtManager := auth.NewJWTManager("test-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewAuthService(userRepo, jwtManager)

	cleanup := func() {
		db.Exec("DELETE FROM users WHERE email LIKE 'authtest%@example.com'")
		db.Close()
	}

	return service, cleanup
}

func TestAuthService_Register(t *testing.T) {
	service, cleanup := setupAuthService(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("should register user successfully", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Email:    "authtest1@example.com",
			Username: "authuser1",
			Password: "securepass123",
		}

		user, err := service.Register(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected user ID to be set")
		}

		if user.Email != req.Email {
			t.Errorf("Expected email %s, got %s", req.Email, user.Email)
		}

		if user.PasswordHash == "" {
			t.Error("Expected password hash to be set")
		}

		if user.PasswordHash == req.Password {
			t.Error("Password should be hashed, not stored in plain text")
		}
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Email:    "authtest2@example.com",
			Username: "authuser2",
			Password: "securepass123",
		}

		_, err := service.Register(ctx, req)
		if err != nil {
			t.Fatalf("Failed to register first user: %v", err)
		}

		// Try to register again with same email
		req2 := &models.CreateUserRequest{
			Email:    "authtest2@example.com",
			Username: "authuser3",
			Password: "securepass123",
		}

		_, err = service.Register(ctx, req2)
		if err != pkgerrors.ErrConflict {
			t.Errorf("Expected ErrConflict, got %v", err)
		}
	})

	t.Run("should validate password strength", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Email:    "authtest3@example.com",
			Username: "authuser3",
			Password: "weak",
		}

		_, err := service.Register(ctx, req)
		if err != pkgerrors.ErrValidation {
			t.Errorf("Expected ErrValidation for weak password, got %v", err)
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	service, cleanup := setupAuthService(t)
	defer cleanup()

	ctx := context.Background()

	// Register a user first
	regReq := &models.CreateUserRequest{
		Email:    "authtest4@example.com",
		Username: "authuser4",
		Password: "securepass123",
	}

	_, err := service.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	t.Run("should login successfully with correct credentials", func(t *testing.T) {
		loginReq := &models.LoginRequest{
			Email:    "authtest4@example.com",
			Password: "securepass123",
		}

		response, err := service.Login(ctx, loginReq)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token to be set")
		}

		if response.RefreshToken == "" {
			t.Error("Expected refresh token to be set")
		}

		if response.User.Email != loginReq.Email {
			t.Errorf("Expected email %s, got %s", loginReq.Email, response.User.Email)
		}

		if response.ExpiresIn <= 0 {
			t.Error("Expected expires_in to be positive")
		}
	})

	t.Run("should return error for wrong password", func(t *testing.T) {
		loginReq := &models.LoginRequest{
			Email:    "authtest4@example.com",
			Password: "wrongpassword",
		}

		_, err := service.Login(ctx, loginReq)
		if err != pkgerrors.ErrInvalidCredentials {
			t.Errorf("Expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		loginReq := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "securepass123",
		}

		_, err := service.Login(ctx, loginReq)
		if err != pkgerrors.ErrInvalidCredentials {
			t.Errorf("Expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	service, cleanup := setupAuthService(t)
	defer cleanup()

	ctx := context.Background()

	// Register and login
	regReq := &models.CreateUserRequest{
		Email:    "authtest5@example.com",
		Username: "authuser5",
		Password: "securepass123",
	}

	registeredUser, err := service.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginReq := &models.LoginRequest{
		Email:    "authtest5@example.com",
		Password: "securepass123",
	}

	loginResp, err := service.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	t.Run("should validate access token", func(t *testing.T) {
		userID, err := service.ValidateAccessToken(ctx, loginResp.AccessToken)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if userID != registeredUser.ID {
			t.Errorf("Expected user ID %d, got %d", registeredUser.ID, userID)
		}
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		_, err := service.ValidateAccessToken(ctx, "invalid-token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})
}
