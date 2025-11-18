package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/issue-tracker/internal/auth"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Validate password strength
	if len(req.Password) < 8 {
		return nil, pkgerrors.ErrValidation
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return nil, pkgerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Return response
	response := &models.LoginResponse{
		TokenPair: models.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    15 * 60, // 15 minutes in seconds
		},
		User: *user,
	}

	return response, nil
}

// ValidateAccessToken validates an access token and returns the user ID
func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (int, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify user still exists
	_, err = s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return nil, pkgerrors.ErrUnauthorized
		}
		return nil, err
	}

	// Generate new tokens
	newAccessToken, err := s.jwtManager.GenerateAccessToken(claims.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(claims.UserID)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    15 * 60,
	}, nil
}

// GetCurrentUser gets the current user by ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
