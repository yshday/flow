package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		switch err {
		case pkgerrors.ErrConflict:
			respondError(w, http.StatusConflict, "User already exists")
		case pkgerrors.ErrValidation:
			respondError(w, http.StatusBadRequest, "Validation error: password must be at least 8 characters")
		default:
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if err == pkgerrors.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "Invalid credentials")
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenPair, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	respondJSON(w, http.StatusOK, tokenPair)
}

// GetMe handles getting current user info
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "User not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// SearchUsers handles searching for users
func (h *AuthHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	users, err := h.authService.SearchUsers(r.Context(), query, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// TokenExchange handles exchanging external user info for Flow tokens
// @Summary Exchange external token for Flow tokens
// @Description Exchange external user info (from SSO/OAuth providers like jmember) for Flow access tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.TokenExchangeRequest true "Token Exchange Request"
// @Success 200 {object} models.TokenExchangeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/token-exchange [post]
func (h *AuthHandler) TokenExchange(w http.ResponseWriter, r *http.Request) {
	var req models.TokenExchangeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Provider == "" || req.ExternalID == "" || req.Email == "" || req.Username == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields: provider, external_id, email, username")
		return
	}

	response, err := h.authService.TokenExchange(r.Context(), &req)
	if err != nil {
		switch err {
		case pkgerrors.ErrConflict:
			respondError(w, http.StatusConflict, "User with this email already exists")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to exchange token")
		}
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
		},
	})
}
