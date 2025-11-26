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

// WebhookHandler handles webhook HTTP requests
type WebhookHandler struct {
	webhookService *service.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// Create handles webhook creation
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required")
		return
	}
	if len(req.Events) == 0 {
		respondError(w, http.StatusBadRequest, "At least one event is required")
		return
	}

	webhook, err := h.webhookService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Admin permission required")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Invalid event type")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create webhook")
		return
	}

	respondJSON(w, http.StatusCreated, webhook)
}

// GetByID handles getting a webhook by ID
func (h *WebhookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Webhook not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get webhook")
		return
	}

	respondJSON(w, http.StatusOK, webhook)
}

// List handles listing webhooks for a project
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	webhooks, err := h.webhookService.List(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list webhooks")
		return
	}

	respondJSON(w, http.StatusOK, webhooks)
}

// Update handles updating a webhook
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	var req models.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	webhook, err := h.webhookService.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Webhook not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Admin permission required")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Invalid event type")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update webhook")
		return
	}

	respondJSON(w, http.StatusOK, webhook)
}

// Delete handles deleting a webhook
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	err = h.webhookService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Webhook not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Admin permission required")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete webhook")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDeliveries handles getting webhook delivery logs
func (h *WebhookHandler) GetDeliveries(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	deliveries, err := h.webhookService.GetDeliveries(r.Context(), id, userID, limit)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Webhook not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get deliveries")
		return
	}

	respondJSON(w, http.StatusOK, deliveries)
}

// GetEventTypes handles returning available webhook event types
func (h *WebhookHandler) GetEventTypes(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, models.AllWebhookEvents())
}
