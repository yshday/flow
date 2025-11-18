package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ReactionHandler handles reaction-related HTTP requests
type ReactionHandler struct {
	service *service.ReactionService
}

// NewReactionHandler creates a new ReactionHandler
func NewReactionHandler(service *service.ReactionService) *ReactionHandler {
	return &ReactionHandler{service: service}
}

// getHTTPStatusCode extracts the HTTP status code from an error
func getReactionHTTPStatusCode(err error) int {
	if appErr, ok := err.(*pkgerrors.AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// AddReactionRequest represents the request to add a reaction
type AddReactionRequest struct {
	Emoji string `json:"emoji"`
}

// AddReaction handles adding a reaction to an issue or comment
// @Summary Add reaction
// @Description Add an emoji reaction to an issue or comment (toggles if already exists)
// @Tags reactions
// @Accept json
// @Produce json
// @Param entity_type path string true "Entity type (issue or comment)"
// @Param entity_id path int true "Entity ID"
// @Param request body AddReactionRequest true "Reaction data"
// @Success 200 {object} map[string]interface{} "Reaction added"
// @Success 204 "Reaction removed (toggle off)"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Entity not found"
// @Security BearerAuth
// @Router /reactions/{entity_type}/{entity_id} [post]
func (h *ReactionHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	var req AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reaction, err := h.service.AddReaction(r.Context(), userID, entityType, entityID, req.Emoji)
	if err != nil {
		statusCode := getReactionHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	// If reaction is nil, it means it was removed (toggled off)
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reaction": reaction,
		"message":  "reaction added",
	})
}

// RemoveReaction handles removing a reaction from an issue or comment
// @Summary Remove reaction
// @Description Remove an emoji reaction from an issue or comment
// @Tags reactions
// @Produce json
// @Param entity_type path string true "Entity type (issue or comment)"
// @Param entity_id path int true "Entity ID"
// @Param emoji path string true "Emoji type"
// @Success 204 "Reaction removed"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Reaction not found"
// @Security BearerAuth
// @Router /reactions/{entity_type}/{entity_id}/{emoji} [delete]
func (h *ReactionHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")
	emoji := r.PathValue("emoji")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	err = h.service.RemoveReaction(r.Context(), userID, entityType, entityID, emoji)
	if err != nil {
		statusCode := getReactionHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReactions handles retrieving all reactions for an issue or comment
// @Summary Get reactions
// @Description Get all reactions for an issue or comment
// @Tags reactions
// @Produce json
// @Param entity_type path string true "Entity type (issue or comment)"
// @Param entity_id path int true "Entity ID"
// @Success 200 {object} map[string]interface{} "List of reactions"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Entity not found"
// @Security BearerAuth
// @Router /reactions/{entity_type}/{entity_id} [get]
func (h *ReactionHandler) GetReactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	reactions, err := h.service.GetReactions(r.Context(), userID, entityType, entityID)
	if err != nil {
		statusCode := getReactionHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reactions": reactions,
	})
}

// GetReactionSummary handles retrieving aggregated reaction counts
// @Summary Get reaction summary
// @Description Get aggregated reaction counts for an issue or comment
// @Tags reactions
// @Produce json
// @Param entity_type path string true "Entity type (issue or comment)"
// @Param entity_id path int true "Entity ID"
// @Success 200 {object} models.ReactionSummary "Reaction summary"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Entity not found"
// @Security BearerAuth
// @Router /reactions/{entity_type}/{entity_id}/summary [get]
func (h *ReactionHandler) GetReactionSummary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	entityType := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	summary, err := h.service.GetReactionSummary(r.Context(), userID, entityType, entityID)
	if err != nil {
		statusCode := getReactionHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
