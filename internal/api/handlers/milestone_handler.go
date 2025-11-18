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

// MilestoneHandler handles milestone HTTP requests
type MilestoneHandler struct {
	milestoneService *service.MilestoneService
}

// NewMilestoneHandler creates a new milestone handler
func NewMilestoneHandler(milestoneService *service.MilestoneService) *MilestoneHandler {
	return &MilestoneHandler{
		milestoneService: milestoneService,
	}
}

// Create handles creating a new milestone
func (h *MilestoneHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	milestone, err := h.milestoneService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create milestone")
		return
	}

	respondJSON(w, http.StatusCreated, milestone)
}

// GetByID handles retrieving a milestone by ID
func (h *MilestoneHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid milestone ID")
		return
	}

	// Check if progress is requested
	withProgress := r.URL.Query().Get("with_progress") == "true"

	var milestone *models.Milestone
	if withProgress {
		milestone, err = h.milestoneService.GetWithProgress(r.Context(), id, userID)
	} else {
		milestone, err = h.milestoneService.GetByID(r.Context(), id, userID)
	}

	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Milestone not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve milestone")
		return
	}

	respondJSON(w, http.StatusOK, milestone)
}

// ListByProject handles listing milestones for a project
func (h *MilestoneHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	milestones, err := h.milestoneService.ListByProjectID(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list milestones")
		return
	}

	respondJSON(w, http.StatusOK, milestones)
}

// Update handles updating a milestone
func (h *MilestoneHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid milestone ID")
		return
	}

	var req models.UpdateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	milestone, err := h.milestoneService.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Milestone not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update milestone")
		return
	}

	respondJSON(w, http.StatusOK, milestone)
}

// Delete handles deleting a milestone
func (h *MilestoneHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid milestone ID")
		return
	}

	err = h.milestoneService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Milestone not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete milestone")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
