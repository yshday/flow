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

// TasklistHandler handles tasklist HTTP requests
type TasklistHandler struct {
	tasklistService *service.TasklistService
}

// NewTasklistHandler creates a new tasklist handler
func NewTasklistHandler(tasklistService *service.TasklistService) *TasklistHandler {
	return &TasklistHandler{
		tasklistService: tasklistService,
	}
}

// Create handles tasklist item creation
// @Summary Create a tasklist item
// @Tags Tasklist
// @Accept json
// @Produce json
// @Param issueId path int true "Issue ID"
// @Param request body models.CreateTasklistItemRequest true "Tasklist item data"
// @Success 201 {object} models.TasklistItem
// @Router /issues/{issueId}/tasklist [post]
func (h *TasklistHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	var req models.CreateTasklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	item, err := h.tasklistService.Create(r.Context(), issueID, &req, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist item")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// List handles listing tasklist items for an issue
// @Summary List tasklist items
// @Tags Tasklist
// @Produce json
// @Param issueId path int true "Issue ID"
// @Success 200 {array} models.TasklistItem
// @Router /issues/{issueId}/tasklist [get]
func (h *TasklistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	items, err := h.tasklistService.ListByIssueID(r.Context(), issueID, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist items")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// Update handles updating a tasklist item
// @Summary Update a tasklist item
// @Tags Tasklist
// @Accept json
// @Produce json
// @Param id path int true "Tasklist item ID"
// @Param request body models.UpdateTasklistItemRequest true "Update data"
// @Success 200 {object} models.TasklistItem
// @Router /tasklist/{id} [put]
func (h *TasklistHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tasklist item ID")
		return
	}

	var req models.UpdateTasklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item, err := h.tasklistService.Update(r.Context(), id, &req, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// Delete handles deleting a tasklist item
// @Summary Delete a tasklist item
// @Tags Tasklist
// @Param id path int true "Tasklist item ID"
// @Success 204 "No Content"
// @Router /tasklist/{id} [delete]
func (h *TasklistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tasklist item ID")
		return
	}

	if err := h.tasklistService.Delete(r.Context(), id, userID); err != nil {
		handleServiceError(w, err, "tasklist item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Toggle handles toggling the completion status of a tasklist item
// @Summary Toggle tasklist item completion
// @Tags Tasklist
// @Produce json
// @Param id path int true "Tasklist item ID"
// @Success 200 {object} models.TasklistItem
// @Router /tasklist/{id}/toggle [patch]
func (h *TasklistHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid tasklist item ID")
		return
	}

	item, err := h.tasklistService.Toggle(r.Context(), id, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// Reorder handles reordering tasklist items
// @Summary Reorder tasklist items
// @Tags Tasklist
// @Accept json
// @Param issueId path int true "Issue ID"
// @Param request body models.ReorderTasklistRequest true "New order"
// @Success 204 "No Content"
// @Router /issues/{issueId}/tasklist/reorder [put]
func (h *TasklistHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	var req models.ReorderTasklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.ItemIDs) == 0 {
		respondError(w, http.StatusBadRequest, "Item IDs are required")
		return
	}

	if err := h.tasklistService.Reorder(r.Context(), issueID, &req, userID); err != nil {
		handleServiceError(w, err, "tasklist items")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProgress handles getting the progress of a tasklist
// @Summary Get tasklist progress
// @Tags Tasklist
// @Produce json
// @Param issueId path int true "Issue ID"
// @Success 200 {object} models.TasklistProgress
// @Router /issues/{issueId}/tasklist/progress [get]
func (h *TasklistHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	progress, err := h.tasklistService.GetProgress(r.Context(), issueID, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// BulkCreate handles creating multiple tasklist items
// @Summary Create multiple tasklist items
// @Tags Tasklist
// @Accept json
// @Produce json
// @Param issueId path int true "Issue ID"
// @Param request body models.BulkCreateTasklistRequest true "Tasklist items data"
// @Success 201 {array} models.TasklistItem
// @Router /issues/{issueId}/tasklist/bulk [post]
func (h *TasklistHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	issueIDStr := r.PathValue("issueId")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	var req models.BulkCreateTasklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "Items are required")
		return
	}

	items, err := h.tasklistService.BulkCreate(r.Context(), issueID, &req, userID)
	if err != nil {
		handleServiceError(w, err, "tasklist items")
		return
	}

	respondJSON(w, http.StatusCreated, items)
}

// Helper function to handle service errors consistently
func handleServiceError(w http.ResponseWriter, err error, entity string) {
	if err == pkgerrors.ErrNotFound {
		respondError(w, http.StatusNotFound, entity+" not found")
		return
	}
	if err == pkgerrors.ErrForbidden {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}
	respondError(w, http.StatusInternalServerError, "Failed to process "+entity)
}
