package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/issue-tracker/internal/auth"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/internal/service"
	"github.com/yourusername/issue-tracker/pkg/database"
)

func setupAuthHandler(t *testing.T) (*AuthHandler, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	jwtManager := auth.NewJWTManager("test-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)
	authService := service.NewAuthService(userRepo, jwtManager)
	handler := NewAuthHandler(authService)

	cleanup := func() {
		db.Exec("DELETE FROM users WHERE email LIKE 'handlertest%@example.com'")
		db.Close()
	}

	return handler, cleanup
}

func TestAuthHandler_Register(t *testing.T) {
	handler, cleanup := setupAuthHandler(t)
	defer cleanup()

	t.Run("should register user successfully", func(t *testing.T) {
		reqBody := models.CreateUserRequest{
			Email:    "handlertest1@example.com",
			Username: "handleruser1",
			Password: "securepass123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var user models.User
		json.NewDecoder(w.Body).Decode(&user)

		if user.Email != reqBody.Email {
			t.Errorf("Expected email %s, got %s", reqBody.Email, user.Email)
		}
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		reqBody := models.CreateUserRequest{
			Email:    "handlertest2@example.com",
			Username: "handleruser2",
			Password: "securepass123",
		}

		body, _ := json.Marshal(reqBody)

		// First registration
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.Register(w, req)

		// Second registration with same email
		body, _ = json.Marshal(reqBody)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		w = httptest.NewRecorder()
		handler.Register(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	handler, cleanup := setupAuthHandler(t)
	defer cleanup()

	// Register a user first
	regBody := models.CreateUserRequest{
		Email:    "handlertest3@example.com",
		Username: "handleruser3",
		Password: "securepass123",
	}

	body, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.Register(w, req)

	t.Run("should login successfully", func(t *testing.T) {
		loginBody := models.LoginRequest{
			Email:    "handlertest3@example.com",
			Password: "securepass123",
		}

		body, _ := json.Marshal(loginBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.LoginResponse
		json.NewDecoder(w.Body).Decode(&response)

		if response.AccessToken == "" {
			t.Error("Expected access token to be set")
		}

		if response.RefreshToken == "" {
			t.Error("Expected refresh token to be set")
		}
	})

	t.Run("should return error for wrong password", func(t *testing.T) {
		loginBody := models.LoginRequest{
			Email:    "handlertest3@example.com",
			Password: "wrongpassword",
		}

		body, _ := json.Marshal(loginBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
