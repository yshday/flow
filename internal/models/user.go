package models

import "time"

// User represents a user in the system
type User struct {
	ID               int       `json:"id"`
	Email            string    `json:"email"`
	Username         string    `json:"username"`
	Name             *string   `json:"name,omitempty"`
	PasswordHash     string    `json:"-"` // Never expose password hash in JSON
	AvatarURL        *string   `json:"avatar_url,omitempty"`
	ExternalID       *string   `json:"external_id,omitempty"`
	ExternalProvider *string   `json:"external_provider,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// LoginResponse represents the login response
type LoginResponse struct {
	TokenPair
	User User `json:"user"`
}

// TokenExchangeRequest represents the request to exchange an external token for Flow tokens
type TokenExchangeRequest struct {
	Provider   string  `json:"provider" validate:"required"`   // e.g., "jmember"
	ExternalID string  `json:"external_id" validate:"required"` // User ID from external system
	Email      string  `json:"email" validate:"required,email"`
	Username   string  `json:"username" validate:"required"`
	Name       *string `json:"name,omitempty"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
}

// TokenExchangeResponse represents the response from token exchange
type TokenExchangeResponse struct {
	TokenPair
	User    User `json:"user"`
	Created bool `json:"created"` // true if user was created, false if existing
}
