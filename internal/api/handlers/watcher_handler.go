package handlers

import (
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// WatcherHandler handles issue watcher HTTP requests
type WatcherHandler struct {
	watcherService *service.WatcherService
}

// NewWatcherHandler creates a new watcher handler
func NewWatcherHandler(watcherService *service.WatcherService) *WatcherHandler {
	return &WatcherHandler{
		watcherService: watcherService,
	}
}

// getWatcherHTTPStatusCode extracts the HTTP status code from an error
func getWatcherHTTPStatusCode(err error) int {
	if appErr, ok := err.(*pkgerrors.AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// WatchIssue subscribes the current user to an issue
// @Summary Watch an issue
// @Description Subscribe to an issue to receive notifications about updates
// @Tags watchers
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param issueNumber path int true "Issue Number"
// @Success 204 "Successfully watching issue"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{projectId}/issues/{issueNumber}/watch [post]
func (h *WatcherHandler) WatchIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	projectIDStr := r.PathValue("projectId")
	issueNumberStr := r.PathValue("issueNumber")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid issue number")
		return
	}

	if err := h.watcherService.WatchIssue(r.Context(), userID, projectID, issueNumber); err != nil {
		statusCode := getWatcherHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnwatchIssue unsubscribes the current user from an issue
// @Summary Unwatch an issue
// @Description Unsubscribe from an issue to stop receiving notifications
// @Tags watchers
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param issueNumber path int true "Issue Number"
// @Success 204 "Successfully unwatched issue"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{projectId}/issues/{issueNumber}/watch [delete]
func (h *WatcherHandler) UnwatchIssue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	projectIDStr := r.PathValue("projectId")
	issueNumberStr := r.PathValue("issueNumber")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid issue number")
		return
	}

	if err := h.watcherService.UnwatchIssue(r.Context(), userID, projectID, issueNumber); err != nil {
		statusCode := getWatcherHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckWatchingStatus checks if the current user is watching an issue
// @Summary Check if watching an issue
// @Description Check if the current user is subscribed to an issue
// @Tags watchers
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param issueNumber path int true "Issue Number"
// @Success 200 {object} map[string]bool "watching status"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{projectId}/issues/{issueNumber}/watching [get]
func (h *WatcherHandler) CheckWatchingStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	projectIDStr := r.PathValue("projectId")
	issueNumberStr := r.PathValue("issueNumber")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid issue number")
		return
	}

	isWatching, err := h.watcherService.IsWatchingIssue(r.Context(), userID, projectID, issueNumber)
	if err != nil {
		statusCode := getWatcherHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"watching": isWatching})
}

// GetWatchers retrieves all watchers for an issue
// @Summary Get issue watchers
// @Description Get list of users watching an issue
// @Tags watchers
// @Security BearerAuth
// @Param projectId path int true "Project ID"
// @Param issueNumber path int true "Issue Number"
// @Success 200 {object} []models.IssueWatcher
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{projectId}/issues/{issueNumber}/watchers [get]
func (h *WatcherHandler) GetWatchers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	projectIDStr := r.PathValue("projectId")
	issueNumberStr := r.PathValue("issueNumber")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid issue number")
		return
	}

	watchers, err := h.watcherService.GetWatchersForIssue(r.Context(), userID, projectID, issueNumber)
	if err != nil {
		statusCode := getWatcherHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, watchers)
}

// GetWatchedIssues retrieves all issues the current user is watching
// @Summary Get watched issues
// @Description Get list of issues the current user is watching
// @Tags watchers
// @Security BearerAuth
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} []models.IssueWatcher
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/watching [get]
func (h *WatcherHandler) GetWatchedIssues(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	watchers, err := h.watcherService.GetWatchedIssuesForUser(r.Context(), userID, limit, offset)
	if err != nil {
		statusCode := getWatcherHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, watchers)
}
