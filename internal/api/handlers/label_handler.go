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

// LabelHandler handles label HTTP requests
type LabelHandler struct {
	labelService *service.LabelService
}

// NewLabelHandler creates a new label handler
func NewLabelHandler(labelService *service.LabelService) *LabelHandler {
	return &LabelHandler{
		labelService: labelService,
	}
}

// Create handles label creation
func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Color == "" {
		respondError(w, http.StatusBadRequest, "Color is required")
		return
	}

	label, err := h.labelService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create label")
		return
	}

	respondJSON(w, http.StatusCreated, label)
}

// List handles listing labels for a project
func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	labels, err := h.labelService.List(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list labels")
		return
	}

	respondJSON(w, http.StatusOK, labels)
}

// GetByID handles getting a label by ID
func (h *LabelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid label ID")
		return
	}

	label, err := h.labelService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Label not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get label")
		return
	}

	respondJSON(w, http.StatusOK, label)
}

// Update handles updating a label
func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid label ID")
		return
	}

	var req models.UpdateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	label, err := h.labelService.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Label not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update label")
		return
	}

	respondJSON(w, http.StatusOK, label)
}

// Delete handles deleting a label
func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid label ID")
		return
	}

	err = h.labelService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Label not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete label")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddToIssue handles adding a label to an issue
func (h *LabelHandler) AddToIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	labelIDStr := r.PathValue("labelId")
	labelID, err := strconv.Atoi(labelIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid label ID")
		return
	}

	err = h.labelService.AddToIssue(r.Context(), issueID, labelID, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue or label not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Label must belong to the same project as the issue")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to add label to issue")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveFromIssue handles removing a label from an issue
func (h *LabelHandler) RemoveFromIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	labelIDStr := r.PathValue("labelId")
	labelID, err := strconv.Atoi(labelIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid label ID")
		return
	}

	err = h.labelService.RemoveFromIssue(r.Context(), issueID, labelID, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue or label not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove label from issue")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByIssueID handles listing labels for an issue
func (h *LabelHandler) ListByIssueID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	labels, err := h.labelService.ListByIssueID(r.Context(), issueID, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list labels")
		return
	}

	respondJSON(w, http.StatusOK, labels)
}
