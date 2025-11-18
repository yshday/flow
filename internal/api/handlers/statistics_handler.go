package handlers

import (
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// StatisticsHandler handles statistics HTTP requests
type StatisticsHandler struct {
	statisticsService *service.StatisticsService
}

// NewStatisticsHandler creates a new statistics handler
func NewStatisticsHandler(statisticsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

// GetProjectStatistics handles retrieving project statistics
func (h *StatisticsHandler) GetProjectStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	stats, err := h.statisticsService.GetProjectStatistics(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve project statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetIssueStatistics handles retrieving issue statistics for a project
func (h *StatisticsHandler) GetIssueStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	stats, err := h.statisticsService.GetIssueStatistics(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve issue statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetUserActivityStatistics handles retrieving user activity statistics
func (h *StatisticsHandler) GetUserActivityStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	stats, err := h.statisticsService.GetUserActivityStatistics(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve user activity statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
