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

// IntegrationHandler handles integration HTTP requests
type IntegrationHandler struct {
	integrationService *service.IntegrationService
}

// NewIntegrationHandler creates a new integration handler
func NewIntegrationHandler(integrationService *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
	}
}

// Create handles integration creation
func (h *IntegrationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Type == "" {
		respondError(w, http.StatusBadRequest, "Type is required")
		return
	}
	if req.WebhookURL == "" {
		respondError(w, http.StatusBadRequest, "Webhook URL is required")
		return
	}
	if len(req.Events) == 0 {
		respondError(w, http.StatusBadRequest, "At least one event is required")
		return
	}

	integration, err := h.integrationService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Admin permission required")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Invalid type or event")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create integration")
		return
	}

	respondJSON(w, http.StatusCreated, integration)
}

// GetByID handles getting an integration by ID
func (h *IntegrationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	integration, err := h.integrationService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Integration not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get integration")
		return
	}

	respondJSON(w, http.StatusOK, integration)
}

// List handles listing integrations for a project
func (h *IntegrationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	integrations, err := h.integrationService.List(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list integrations")
		return
	}

	respondJSON(w, http.StatusOK, integrations)
}

// Update handles updating an integration
func (h *IntegrationHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	var req models.UpdateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	integration, err := h.integrationService.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Integration not found")
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
		respondError(w, http.StatusInternalServerError, "Failed to update integration")
		return
	}

	respondJSON(w, http.StatusOK, integration)
}

// Delete handles deleting an integration
func (h *IntegrationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	err = h.integrationService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Integration not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Admin permission required")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete integration")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMessages handles getting integration message logs
func (h *IntegrationHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages, err := h.integrationService.GetMessages(r.Context(), id, userID, limit)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Integration not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

// GetIntegrationTypes handles returning available integration types
func (h *IntegrationHandler) GetIntegrationTypes(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, models.AllIntegrationTypes())
}

// TestIntegration handles testing an integration by sending a test message
func (h *IntegrationHandler) TestIntegration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid integration ID")
		return
	}

	integration, err := h.integrationService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Integration not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get integration")
		return
	}

	// Create a test issue to send
	testIssue := &models.Issue{
		ID:          0,
		ProjectID:   integration.ProjectID,
		IssueNumber: 0,
		Title:       "Test Notification",
		Description: stringPtr("This is a test notification from Flow Issue Tracker."),
		Status:      models.IssueStatusOpen,
		Priority:    models.PriorityMedium,
	}

	// Send test event
	err = h.integrationService.SendEvent(r.Context(), integration.ProjectID, models.EventIssueCreated, testIssue)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send test message")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Test message sent"})
}

func stringPtr(s string) *string {
	return &s
}
