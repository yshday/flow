package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndValidateAccessToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)

	userID := 123

	t.Run("should generate and validate access token", func(t *testing.T) {
		token, err := jwtManager.GenerateAccessToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected user ID %d, got %d", userID, claims.UserID)
		}

		if claims.Type != "access" {
			t.Errorf("Expected token type 'access', got '%s'", claims.Type)
		}
	})

	t.Run("should reject invalid access token", func(t *testing.T) {
		_, err := jwtManager.ValidateAccessToken("invalid-token")
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("should reject refresh token as access token", func(t *testing.T) {
		refreshToken, err := jwtManager.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		_, err = jwtManager.ValidateAccessToken(refreshToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestJWTManager_GenerateAndValidateRefreshToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)

	userID := 456

	t.Run("should generate and validate refresh token", func(t *testing.T) {
		token, err := jwtManager.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		claims, err := jwtManager.ValidateRefreshToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected user ID %d, got %d", userID, claims.UserID)
		}

		if claims.Type != "refresh" {
			t.Errorf("Expected token type 'refresh', got '%s'", claims.Type)
		}
	})

	t.Run("should reject access token as refresh token", func(t *testing.T) {
		accessToken, err := jwtManager.GenerateAccessToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		_, err = jwtManager.ValidateRefreshToken(accessToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	// Create manager with very short TTL for testing
	jwtManager := NewJWTManager("test-secret", "test-refresh-secret", 1*time.Second, 2*time.Second)

	userID := 789

	t.Run("should reject expired access token", func(t *testing.T) {
		token, err := jwtManager.GenerateAccessToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		_, err = jwtManager.ValidateAccessToken(token)
		if err != ErrExpiredToken {
			t.Errorf("Expected ErrExpiredToken, got %v", err)
		}
	})
}
