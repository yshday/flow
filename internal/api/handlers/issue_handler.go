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

// IssueHandler handles issue HTTP requests
type IssueHandler struct {
	issueService *service.IssueService
}

// NewIssueHandler creates a new issue handler
func NewIssueHandler(issueService *service.IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
	}
}

// Create handles issue creation
func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	issue, err := h.issueService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Invalid issue type hierarchy")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusBadRequest, "Parent issue or epic not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create issue")
		return
	}

	respondJSON(w, http.StatusCreated, issue)
}

// GetByID handles getting an issue by ID
func (h *IssueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	issue, err := h.issueService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get issue")
		return
	}

	respondJSON(w, http.StatusOK, issue)
}

// GetByNumber handles getting an issue by project ID and issue number
func (h *IssueHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	issueNumberStr := r.PathValue("issueNumber")
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue number")
		return
	}

	issue, err := h.issueService.GetByNumber(r.Context(), projectID, issueNumber, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get issue")
		return
	}

	respondJSON(w, http.StatusOK, issue)
}

// GetByProjectKey handles getting an issue by project key and issue number
func (h *IssueHandler) GetByProjectKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectKey := r.PathValue("projectKey")
	issueNumberStr := r.PathValue("issueNumber")
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue number")
		return
	}

	issue, err := h.issueService.GetByProjectKey(r.Context(), projectKey, issueNumber, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get issue")
		return
	}

	respondJSON(w, http.StatusOK, issue)
}

// List handles listing issues for a project
func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Parse query parameters for filtering
	filter := &models.IssueFilter{
		ProjectID: projectID,
	}

	// Status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.IssueStatus(statusStr)
		filter.Status = &status
	}

	// Priority filter
	if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
		priority := models.IssuePriority(priorityStr)
		filter.Priority = &priority
	}

	// Assignee filter
	if assigneeStr := r.URL.Query().Get("assignee_id"); assigneeStr != "" {
		assigneeID, err := strconv.Atoi(assigneeStr)
		if err == nil {
			filter.AssigneeID = &assigneeID
		}
	}

	// Milestone filter
	if milestoneStr := r.URL.Query().Get("milestone_id"); milestoneStr != "" {
		milestoneID, err := strconv.Atoi(milestoneStr)
		if err == nil {
			filter.MilestoneID = &milestoneID
		}
	}

	// Label filter
	if labelStr := r.URL.Query().Get("label_id"); labelStr != "" {
		labelID, err := strconv.Atoi(labelStr)
		if err == nil {
			filter.LabelIDs = []int{labelID}
		}
	}

	// Issue type filter
	if issueTypeStr := r.URL.Query().Get("issue_type"); issueTypeStr != "" {
		issueType := models.IssueType(issueTypeStr)
		filter.IssueType = &issueType
	}

	// Parent issue filter
	if parentStr := r.URL.Query().Get("parent_issue_id"); parentStr != "" {
		parentID, err := strconv.Atoi(parentStr)
		if err == nil {
			filter.ParentIssueID = &parentID
		}
	}

	// Epic filter
	if epicStr := r.URL.Query().Get("epic_id"); epicStr != "" {
		epicID, err := strconv.Atoi(epicStr)
		if err == nil {
			filter.EpicID = &epicID
		}
	}

	// HasParent filter (for subtasks or top-level only)
	if hasParentStr := r.URL.Query().Get("has_parent"); hasParentStr != "" {
		hasParent := hasParentStr == "true"
		filter.HasParent = &hasParent
	}

	// Search filter (support both 'q' and 'search')
	if search := r.URL.Query().Get("q"); search != "" {
		filter.Search = search
	} else if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = search
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	issues, err := h.issueService.List(r.Context(), filter, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list issues")
		return
	}

	respondJSON(w, http.StatusOK, issues)
}

// Update handles updating an issue
func (h *IssueHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	var req models.UpdateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	issue, err := h.issueService.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrConflict {
			respondError(w, http.StatusConflict, "Issue was modified by another user")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Invalid issue type hierarchy")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update issue")
		return
	}

	respondJSON(w, http.StatusOK, issue)
}

// Delete handles deleting an issue
func (h *IssueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	err = h.issueService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete issue")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MoveToColumn handles moving an issue to a different board column
func (h *IssueHandler) MoveToColumn(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	var req models.MoveIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	issue, err := h.issueService.MoveToColumn(r.Context(), id, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue or column not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrConflict {
			respondError(w, http.StatusConflict, "Issue was modified by another user")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Column must belong to the same project as the issue")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to move issue")
		return
	}

	respondJSON(w, http.StatusOK, issue)
}

// GetSubtasks handles getting subtasks for an issue
func (h *IssueHandler) GetSubtasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	subtasks, err := h.issueService.GetSubtasks(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subtasks")
		return
	}

	respondJSON(w, http.StatusOK, subtasks)
}

// GetEpicIssues handles getting issues under an epic
func (h *IssueHandler) GetEpicIssues(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid epic ID")
		return
	}

	issues, err := h.issueService.GetEpicIssues(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Epic not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Issue is not an epic")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get epic issues")
		return
	}

	respondJSON(w, http.StatusOK, issues)
}

// GetEpics handles getting all epics for a project
func (h *IssueHandler) GetEpics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	epics, err := h.issueService.GetEpics(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get epics")
		return
	}

	respondJSON(w, http.StatusOK, epics)
}

// SubtaskProgressResponse represents subtask progress data
type SubtaskProgressResponse struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
}

// GetSubtaskProgress handles getting subtask progress for an issue
func (h *IssueHandler) GetSubtaskProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	total, completed, err := h.issueService.GetSubtaskProgress(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subtask progress")
		return
	}

	respondJSON(w, http.StatusOK, SubtaskProgressResponse{
		Total:     total,
		Completed: completed,
	})
}

// GetEpicProgress handles getting issue completion progress for an epic
func (h *IssueHandler) GetEpicProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid epic ID")
		return
	}

	total, completed, err := h.issueService.GetEpicProgress(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Epic not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Issue is not an epic")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get epic progress")
		return
	}

	respondJSON(w, http.StatusOK, SubtaskProgressResponse{
		Total:     total,
		Completed: completed,
	})
}
