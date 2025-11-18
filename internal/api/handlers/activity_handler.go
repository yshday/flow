package handlers

import (
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ActivityHandler handles activity HTTP requests
type ActivityHandler struct {
	activityService *service.ActivityService
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// ListByProject handles listing activities for a project
func (h *ActivityHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Parse pagination parameters
	limit, offset := parsePagination(r)

	activities, err := h.activityService.ListByProjectID(r.Context(), projectID, limit, offset, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list activities")
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// ListByIssue handles listing activities for an issue
func (h *ActivityHandler) ListByIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	// Parse pagination parameters
	limit, offset := parsePagination(r)

	activities, err := h.activityService.ListByIssueID(r.Context(), issueID, limit, offset, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list activities")
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// parsePagination extracts limit and offset from query parameters
func parsePagination(r *http.Request) (limit, offset int) {
	limit = 50 // default
	offset = 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
