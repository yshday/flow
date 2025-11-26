package service

import (
	"context"
	"log/slog"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// TasklistService handles tasklist business logic
type TasklistService struct {
	tasklistRepo    *repository.TasklistRepository
	issueRepo       *repository.IssueRepository
	authService     *AuthorizationService
	activityService *ActivityService
}

// NewTasklistService creates a new tasklist service
func NewTasklistService(
	tasklistRepo *repository.TasklistRepository,
	issueRepo *repository.IssueRepository,
	authService *AuthorizationService,
	activityService *ActivityService,
) *TasklistService {
	return &TasklistService{
		tasklistRepo:    tasklistRepo,
		issueRepo:       issueRepo,
		authService:     authService,
		activityService: activityService,
	}
}

// Create creates a new tasklist item
func (s *TasklistService) Create(ctx context.Context, issueID int, req *models.CreateTasklistItemRequest, userID int) (*models.TasklistItem, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	item := &models.TasklistItem{
		IssueID:     issueID,
		Content:     req.Content,
		IsCompleted: false,
	}

	if req.Position != nil {
		item.Position = *req.Position
	}

	created, err := s.tasklistRepo.Create(ctx, item)
	if err != nil {
		slog.Error("failed to create tasklist item", "error", err, "issue_id", issueID)
		return nil, err
	}

	// Log activity
	if s.activityService != nil {
		_, _ = s.activityService.LogActivity(ctx, &models.CreateActivityRequest{
			IssueID:    &issueID,
			UserID:     userID,
			Action:     "tasklist_item_added",
			EntityType: "tasklist_item",
			NewValue:   strPtr(req.Content),
		})
	}

	slog.Info("tasklist item created", "id", created.ID, "issue_id", issueID, "user_id", userID)
	return created, nil
}

// GetByID retrieves a tasklist item by ID
func (s *TasklistService) GetByID(ctx context.Context, id int, userID int) (*models.TasklistItem, error) {
	item, err := s.tasklistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has read permission
	issue, err := s.issueRepo.GetByID(ctx, item.IssueID)
	if err != nil {
		return nil, err
	}

	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return item, nil
}

// ListByIssueID retrieves all tasklist items for an issue
func (s *TasklistService) ListByIssueID(ctx context.Context, issueID int, userID int) ([]*models.TasklistItem, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has read permission
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return s.tasklistRepo.ListByIssueID(ctx, issueID)
}

// Update updates a tasklist item
func (s *TasklistService) Update(ctx context.Context, id int, req *models.UpdateTasklistItemRequest, userID int) (*models.TasklistItem, error) {
	// Get existing item
	item, err := s.tasklistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission
	issue, err := s.issueRepo.GetByID(ctx, item.IssueID)
	if err != nil {
		return nil, err
	}

	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Content != nil {
		item.Content = *req.Content
	}
	if req.Position != nil {
		item.Position = *req.Position
	}
	if req.IsCompleted != nil {
		item.IsCompleted = *req.IsCompleted
	}

	updated, err := s.tasklistRepo.Update(ctx, item)
	if err != nil {
		slog.Error("failed to update tasklist item", "error", err, "id", id)
		return nil, err
	}

	slog.Info("tasklist item updated", "id", id, "user_id", userID)
	return updated, nil
}

// Delete removes a tasklist item
func (s *TasklistService) Delete(ctx context.Context, id int, userID int) error {
	// Get existing item
	item, err := s.tasklistRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has write permission
	issue, err := s.issueRepo.GetByID(ctx, item.IssueID)
	if err != nil {
		return err
	}

	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	if err := s.tasklistRepo.Delete(ctx, id); err != nil {
		slog.Error("failed to delete tasklist item", "error", err, "id", id)
		return err
	}

	// Log activity
	if s.activityService != nil {
		issueID := item.IssueID
		_, _ = s.activityService.LogActivity(ctx, &models.CreateActivityRequest{
			IssueID:    &issueID,
			UserID:     userID,
			Action:     "tasklist_item_removed",
			EntityType: "tasklist_item",
			OldValue:   strPtr(item.Content),
		})
	}

	slog.Info("tasklist item deleted", "id", id, "user_id", userID)
	return nil
}

// Toggle toggles the completion status of a tasklist item
func (s *TasklistService) Toggle(ctx context.Context, id int, userID int) (*models.TasklistItem, error) {
	// Get existing item
	item, err := s.tasklistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission
	issue, err := s.issueRepo.GetByID(ctx, item.IssueID)
	if err != nil {
		return nil, err
	}

	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	toggled, err := s.tasklistRepo.Toggle(ctx, id, userID)
	if err != nil {
		slog.Error("failed to toggle tasklist item", "error", err, "id", id)
		return nil, err
	}

	// Log activity
	if s.activityService != nil {
		action := "tasklist_item_completed"
		if !toggled.IsCompleted {
			action = "tasklist_item_uncompleted"
		}
		issueID := item.IssueID
		_, _ = s.activityService.LogActivity(ctx, &models.CreateActivityRequest{
			IssueID:    &issueID,
			UserID:     userID,
			Action:     action,
			EntityType: "tasklist_item",
			NewValue:   strPtr(item.Content),
		})
	}

	slog.Info("tasklist item toggled", "id", id, "completed", toggled.IsCompleted, "user_id", userID)
	return toggled, nil
}

// Reorder updates positions of tasklist items
func (s *TasklistService) Reorder(ctx context.Context, issueID int, req *models.ReorderTasklistRequest, userID int) error {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return err
	}

	// Check if user has write permission
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	// Validate that all items belong to this issue
	existingItems, err := s.tasklistRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		return err
	}

	existingIDs := make(map[int]bool)
	for _, item := range existingItems {
		existingIDs[item.ID] = true
	}

	for _, id := range req.ItemIDs {
		if !existingIDs[id] {
			return pkgerrors.ErrForbidden
		}
	}

	if err := s.tasklistRepo.Reorder(ctx, issueID, req.ItemIDs); err != nil {
		slog.Error("failed to reorder tasklist items", "error", err, "issue_id", issueID)
		return err
	}

	slog.Info("tasklist items reordered", "issue_id", issueID, "user_id", userID)
	return nil
}

// GetProgress returns the completion progress of a tasklist
func (s *TasklistService) GetProgress(ctx context.Context, issueID int, userID int) (*models.TasklistProgress, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has read permission
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return s.tasklistRepo.GetProgress(ctx, issueID)
}

// BulkCreate creates multiple tasklist items
func (s *TasklistService) BulkCreate(ctx context.Context, issueID int, req *models.BulkCreateTasklistRequest, userID int) ([]*models.TasklistItem, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	created, err := s.tasklistRepo.BulkCreate(ctx, issueID, req.Items)
	if err != nil {
		slog.Error("failed to bulk create tasklist items", "error", err, "issue_id", issueID)
		return nil, err
	}

	slog.Info("tasklist items bulk created", "issue_id", issueID, "count", len(created), "user_id", userID)
	return created, nil
}

// Helper function for string pointer
func strPtr(s string) *string {
	return &s
}
