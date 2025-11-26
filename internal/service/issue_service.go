package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
	"github.com/yourusername/issue-tracker/pkg/markdown"
)

// IssueService handles issue business logic
type IssueService struct {
	issueRepo          *repository.IssueRepository
	watcherRepo        *repository.IssueWatcherRepository
	authService        *AuthorizationService
	db                 *sql.DB
	cache              pkgcache.Cache
	markdownRenderer   *markdown.Renderer
	mentionService     *MentionService
	referenceService   *IssueReferenceService
	webhookService     *WebhookService
	integrationService *IntegrationService
}

// NewIssueService creates a new issue service
func NewIssueService(issueRepo *repository.IssueRepository, watcherRepo *repository.IssueWatcherRepository, authService *AuthorizationService, db *sql.DB, cache pkgcache.Cache, mdRenderer *markdown.Renderer, mentionService *MentionService, referenceService *IssueReferenceService, webhookService *WebhookService, integrationService *IntegrationService) *IssueService {
	return &IssueService{
		issueRepo:          issueRepo,
		watcherRepo:        watcherRepo,
		authService:        authService,
		db:                 db,
		cache:              cache,
		markdownRenderer:   mdRenderer,
		mentionService:     mentionService,
		referenceService:   referenceService,
		webhookService:     webhookService,
		integrationService: integrationService,
	}
}

// Create creates a new issue
func (s *IssueService) Create(ctx context.Context, projectID int, req *models.CreateIssueRequest, userID int) (*models.Issue, error) {
	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Set default values
	priority := models.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	issueType := models.IssueTypeTask
	if req.IssueType != nil {
		issueType = *req.IssueType
	}

	// Validate issue type hierarchy rules
	if err := s.validateIssueTypeHierarchy(ctx, issueType, req.ParentIssueID, req.EpicID, projectID); err != nil {
		return nil, err
	}

	issue := &models.Issue{
		ProjectID:     projectID,
		Title:         req.Title,
		Description:   req.Description,
		Status:        models.IssueStatusOpen,
		Priority:      priority,
		IssueType:     issueType,
		ParentIssueID: req.ParentIssueID,
		EpicID:        req.EpicID,
		AssigneeID:    req.AssigneeID,
		ReporterID:    userID,
		ColumnID:      req.ColumnID,
		MilestoneID:   req.MilestoneID,
	}

	// Render markdown description to HTML
	if req.Description != nil && *req.Description != "" {
		html := s.markdownRenderer.RenderToHTML(*req.Description)
		issue.DescriptionHTML = &html
	}

	created, err := s.issueRepo.Create(ctx, issue)
	if err != nil {
		return nil, err
	}

	// Process @mentions in description
	if req.Description != nil && *req.Description != "" {
		mentionedUserIDs, err := s.mentionService.ProcessMentions(ctx, *req.Description, "issue", created.ID, userID)
		if err != nil {
			// Log error but don't fail the whole operation
			// TODO: Add proper logging
		} else if len(mentionedUserIDs) > 0 {
			// Create notifications for mentioned users
			_ = s.mentionService.CreateNotificationsForMentions(ctx, mentionedUserIDs, "issue", created.ID, userID)
		}

		// Process #issue_number references
		_, _ = s.referenceService.ProcessReferences(ctx, *req.Description, "issue", created.ID, projectID)
	}

	// Automatically make the creator watch this issue
	_ = s.watcherRepo.Watch(ctx, userID, created.ID)

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, projectID)

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), projectID, models.EventIssueCreated, userID, created)
	}

	// Send integration notifications (Slack, Discord, etc.)
	if s.integrationService != nil {
		go s.integrationService.SendEvent(context.Background(), projectID, models.EventIssueCreated, created)
	}

	return created, nil
}

// GetByID retrieves an issue by ID
func (s *IssueService) GetByID(ctx context.Context, id int, userID int) (*models.Issue, error) {
	// Get issue first to know which project it belongs to
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	hasAccess, err := s.userHasAccess(ctx, userID, issue.ProjectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return issue, nil
}

// GetByNumber retrieves an issue by project ID and issue number
func (s *IssueService) GetByNumber(ctx context.Context, projectID int, issueNumber int, userID int) (*models.Issue, error) {
	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
}

// GetByProjectKey retrieves an issue by project key and issue number
func (s *IssueService) GetByProjectKey(ctx context.Context, projectKey string, issueNumber int, userID int) (*models.Issue, error) {
	// Get project by key
	var projectID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM projects WHERE key = $1", projectKey).Scan(&projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
}

// List retrieves issues with filtering
func (s *IssueService) List(ctx context.Context, filter *models.IssueFilter, userID int) ([]*models.Issue, error) {
	// Check if user has access to the project
	hasAccess, err := s.userHasAccess(ctx, userID, filter.ProjectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.List(ctx, filter)
}

// Update updates an issue
func (s *IssueService) Update(ctx context.Context, id int, req *models.UpdateIssueRequest, userID int) (*models.Issue, error) {
	// Get existing issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Title != nil {
		issue.Title = *req.Title
	}
	if req.Description != nil {
		issue.Description = req.Description
		// Render markdown description to HTML
		if *req.Description != "" {
			html := s.markdownRenderer.RenderToHTML(*req.Description)
			issue.DescriptionHTML = &html
		} else {
			issue.DescriptionHTML = nil
		}
	}
	if req.Status != nil {
		issue.Status = *req.Status
	}
	if req.Priority != nil {
		issue.Priority = *req.Priority
	}
	if req.IssueType != nil {
		// Validate type change
		if err := s.validateIssueTypeHierarchy(ctx, *req.IssueType, issue.ParentIssueID, req.EpicID, issue.ProjectID); err != nil {
			return nil, err
		}
		issue.IssueType = *req.IssueType
	}
	if req.EpicID != nil {
		// Validate epic change
		if err := s.validateIssueTypeHierarchy(ctx, issue.IssueType, issue.ParentIssueID, req.EpicID, issue.ProjectID); err != nil {
			return nil, err
		}
		issue.EpicID = req.EpicID
	}
	if req.AssigneeID != nil {
		issue.AssigneeID = req.AssigneeID
	}
	if req.MilestoneID != nil {
		issue.MilestoneID = req.MilestoneID
	}

	// Save changes
	err = s.issueRepo.Update(ctx, issue)
	if err != nil {
		return nil, err
	}

	// Handle mentions and references if description was updated
	if req.Description != nil {
		// Delete old mentions and references for this issue
		_ = s.mentionService.mentionRepo.DeleteByEntity(ctx, "issue", id)
		_ = s.referenceService.referenceRepo.DeleteBySource(ctx, "issue", id)

		// Process new mentions and references
		if *req.Description != "" {
			mentionedUserIDs, err := s.mentionService.ProcessMentions(ctx, *req.Description, "issue", id, userID)
			if err == nil && len(mentionedUserIDs) > 0 {
				// Create notifications for mentioned users
				_ = s.mentionService.CreateNotificationsForMentions(ctx, mentionedUserIDs, "issue", id, userID)
			}

			// Process #issue_number references
			_, _ = s.referenceService.ProcessReferences(ctx, *req.Description, "issue", id, issue.ProjectID)
		}
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, issue.ProjectID)

	// Get updated issue
	updated, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventIssueUpdated, userID, updated)
	}

	// Send integration notifications (Slack, Discord, etc.)
	if s.integrationService != nil {
		go s.integrationService.SendEvent(context.Background(), issue.ProjectID, models.EventIssueUpdated, updated)
	}

	return updated, nil
}

// Delete deletes an issue
func (s *IssueService) Delete(ctx context.Context, id int, userID int) error {
	// Get issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has admin permission (only admins and owners can delete issues)
	if err := s.authService.CheckAdminPermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	// Copy issue data before deletion for webhook
	deletedIssue := *issue

	err = s.issueRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, issue.ProjectID)

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventIssueDeleted, userID, &deletedIssue)
	}

	// Send integration notifications (Slack, Discord, etc.)
	if s.integrationService != nil {
		go s.integrationService.SendEvent(context.Background(), issue.ProjectID, models.EventIssueDeleted, &deletedIssue)
	}

	return nil
}

// MoveToColumn moves an issue to a different board column
func (s *IssueService) MoveToColumn(ctx context.Context, id int, req *models.MoveIssueRequest, userID int) (*models.Issue, error) {
	// Get existing issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	// Verify version for optimistic locking
	if issue.Version != req.Version {
		return nil, pkgerrors.ErrConflict
	}

	// Verify target column exists and belongs to the same project
	var columnProjectID int
	err = s.db.QueryRowContext(ctx, `
		SELECT project_id FROM board_columns WHERE id = $1
	`, req.ColumnID).Scan(&columnProjectID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	if columnProjectID != issue.ProjectID {
		return nil, pkgerrors.ErrValidation
	}

	// Update issue's column
	issue.ColumnID = &req.ColumnID
	issue.ColumnPosition = req.Position

	// Update status if provided (for kanban board auto-status change)
	if req.Status != nil {
		issue.Status = *req.Status
	}

	err = s.issueRepo.Update(ctx, issue)
	if err != nil {
		return nil, err
	}

	// Invalidate project caches (for board view updates)
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, issue.ProjectID)

	// Get updated issue
	updated, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventIssueMoved, userID, updated)
	}

	// Send integration notifications (Slack, Discord, etc.)
	if s.integrationService != nil {
		go s.integrationService.SendEvent(context.Background(), issue.ProjectID, models.EventIssueMoved, updated)
	}

	return updated, nil
}

// userHasAccess checks if user has any access to the project
func (s *IssueService) userHasAccess(ctx context.Context, userID int, projectID int) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// userHasPermission checks if user has specific role in the project
func (s *IssueService) userHasPermission(ctx context.Context, userID int, projectID int, allowedRoles []models.ProjectRole) (bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	for _, allowedRole := range allowedRoles {
		if models.ProjectRole(role) == allowedRole {
			return true, nil
		}
	}

	return false, nil
}

// validateIssueTypeHierarchy validates issue type hierarchy rules
func (s *IssueService) validateIssueTypeHierarchy(ctx context.Context, issueType models.IssueType, parentIssueID *int, epicID *int, projectID int) error {
	// Rule 1: Subtasks must have a parent
	if issueType == models.IssueTypeSubtask && parentIssueID == nil {
		return pkgerrors.ErrValidation // Subtask must have a parent
	}

	// Rule 2: Epics cannot have a parent
	if issueType == models.IssueTypeEpic && parentIssueID != nil {
		return pkgerrors.ErrValidation // Epic cannot have a parent
	}

	// Rule 3: Subtasks cannot have an epic (they inherit from parent)
	if issueType == models.IssueTypeSubtask && epicID != nil {
		return pkgerrors.ErrValidation // Subtask cannot have an epic
	}

	// Rule 4: Epics cannot belong to another epic
	if issueType == models.IssueTypeEpic && epicID != nil {
		return pkgerrors.ErrValidation // Epic cannot belong to an epic
	}

	// Rule 5: If parent is specified, validate it exists and is in the same project
	if parentIssueID != nil {
		parent, err := s.issueRepo.GetByID(ctx, *parentIssueID)
		if err != nil {
			return err
		}
		if parent.ProjectID != projectID {
			return pkgerrors.ErrValidation // Parent must be in the same project
		}
		// Parent cannot be an epic
		if parent.IssueType == models.IssueTypeEpic {
			return pkgerrors.ErrValidation // Cannot add subtask to epic
		}
		// Parent cannot be a subtask (no nesting subtasks)
		if parent.IssueType == models.IssueTypeSubtask {
			return pkgerrors.ErrValidation // Cannot add subtask to a subtask
		}
	}

	// Rule 6: If epic is specified, validate it exists and is actually an epic
	if epicID != nil {
		epic, err := s.issueRepo.GetByID(ctx, *epicID)
		if err != nil {
			return err
		}
		if epic.ProjectID != projectID {
			return pkgerrors.ErrValidation // Epic must be in the same project
		}
		if epic.IssueType != models.IssueTypeEpic {
			return pkgerrors.ErrValidation // Referenced issue must be an epic
		}
	}

	return nil
}

// GetSubtasks retrieves all subtasks for an issue
func (s *IssueService) GetSubtasks(ctx context.Context, issueID int, userID int) ([]*models.Issue, error) {
	// Get parent issue to check access
	parent, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, parent.ProjectID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.GetSubtasks(ctx, issueID)
}

// GetEpicIssues retrieves all issues under an epic
func (s *IssueService) GetEpicIssues(ctx context.Context, epicID int, userID int) ([]*models.Issue, error) {
	// Get epic to check access and type
	epic, err := s.issueRepo.GetByID(ctx, epicID)
	if err != nil {
		return nil, err
	}

	// Verify it's an epic
	if epic.IssueType != models.IssueTypeEpic {
		return nil, pkgerrors.ErrValidation
	}

	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, epic.ProjectID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.GetEpicIssues(ctx, epicID)
}

// GetEpics retrieves all epics for a project
func (s *IssueService) GetEpics(ctx context.Context, projectID int, userID int) ([]*models.Issue, error) {
	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.issueRepo.GetEpics(ctx, projectID)
}

// GetSubtaskProgress returns subtask completion stats for an issue
func (s *IssueService) GetSubtaskProgress(ctx context.Context, issueID int, userID int) (total int, completed int, err error) {
	// Get parent issue to check access
	parent, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return 0, 0, err
	}

	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, parent.ProjectID)
	if err != nil {
		return 0, 0, err
	}
	if !hasAccess {
		return 0, 0, pkgerrors.ErrForbidden
	}

	return s.issueRepo.CountSubtasks(ctx, issueID)
}

// GetEpicProgress returns issue completion stats for an epic
func (s *IssueService) GetEpicProgress(ctx context.Context, epicID int, userID int) (total int, completed int, err error) {
	// Get epic to check access
	epic, err := s.issueRepo.GetByID(ctx, epicID)
	if err != nil {
		return 0, 0, err
	}

	// Verify it's an epic
	if epic.IssueType != models.IssueTypeEpic {
		return 0, 0, pkgerrors.ErrValidation
	}

	// Check access
	hasAccess, err := s.userHasAccess(ctx, userID, epic.ProjectID)
	if err != nil {
		return 0, 0, err
	}
	if !hasAccess {
		return 0, 0, pkgerrors.ErrForbidden
	}

	return s.issueRepo.CountEpicIssues(ctx, epicID)
}
