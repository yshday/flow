package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/service"
)

// SearchHandler handles search HTTP requests
type SearchHandler struct {
	searchService *service.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchIssues handles searching for issues
func (h *SearchHandler) SearchIssues(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Parse query parameters
	query := r.URL.Query().Get("q")

	// Project ID filter
	var projectID *int
	if projectIDStr := r.URL.Query().Get("project_id"); projectIDStr != "" {
		pid, err := strconv.Atoi(projectIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid project ID")
			return
		}
		projectID = &pid
	}

	// Status filter (comma-separated)
	var status []string
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status = strings.Split(statusStr, ",")
	}

	// Priority filter (comma-separated)
	var priority []string
	if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
		priority = strings.Split(priorityStr, ",")
	}

	// Assignee ID filter
	var assigneeID *int
	if assigneeIDStr := r.URL.Query().Get("assignee_id"); assigneeIDStr != "" {
		aid, err := strconv.Atoi(assigneeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid assignee ID")
			return
		}
		assigneeID = &aid
	}

	// Reporter ID filter
	var reporterID *int
	if reporterIDStr := r.URL.Query().Get("reporter_id"); reporterIDStr != "" {
		rid, err := strconv.Atoi(reporterIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid reporter ID")
			return
		}
		reporterID = &rid
	}

	// Label IDs filter (comma-separated)
	var labelIDs []int
	if labelIDsStr := r.URL.Query().Get("label_ids"); labelIDsStr != "" {
		labelIDStrs := strings.Split(labelIDsStr, ",")
		for _, lidStr := range labelIDStrs {
			lid, err := strconv.Atoi(strings.TrimSpace(lidStr))
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid label ID")
				return
			}
			labelIDs = append(labelIDs, lid)
		}
	}

	// Pagination
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	req := &models.IssueSearchRequest{
		Query:      query,
		ProjectID:  projectID,
		Status:     status,
		Priority:   priority,
		AssigneeID: assigneeID,
		ReporterID: reporterID,
		LabelIDs:   labelIDs,
		Limit:      limit,
		Offset:     offset,
	}

	results, err := h.searchService.SearchIssues(r.Context(), req, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search issues")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// SearchProjects handles searching for projects
func (h *SearchHandler) SearchProjects(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Parse query parameters
	query := r.URL.Query().Get("q")

	// Pagination
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	req := &models.ProjectSearchRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}

	results, err := h.searchService.SearchProjects(r.Context(), req, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search projects")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// Search handles unified search across issues and projects
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	searchType := r.URL.Query().Get("type") // "issues", "projects", or "all"
	if searchType == "" {
		searchType = "all"
	}

	// Pagination
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	response := make(map[string]interface{})

	if searchType == "issues" || searchType == "all" {
		issueReq := &models.IssueSearchRequest{
			Query:  query,
			Limit:  limit,
			Offset: offset,
		}
		issueResults, err := h.searchService.SearchIssues(r.Context(), issueReq, userID)
		if err == nil {
			response["issues"] = issueResults
		}
	}

	if searchType == "projects" || searchType == "all" {
		projectReq := &models.ProjectSearchRequest{
			Query:  query,
			Limit:  limit,
			Offset: offset,
		}
		projectResults, err := h.searchService.SearchProjects(r.Context(), projectReq, userID)
		if err == nil {
			response["projects"] = projectResults
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
