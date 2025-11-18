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
	issueRepo        *repository.IssueRepository
	watcherRepo      *repository.IssueWatcherRepository
	authService      *AuthorizationService
	db               *sql.DB
	cache            pkgcache.Cache
	markdownRenderer *markdown.Renderer
	mentionService   *MentionService
	referenceService *IssueReferenceService
}

// NewIssueService creates a new issue service
func NewIssueService(issueRepo *repository.IssueRepository, watcherRepo *repository.IssueWatcherRepository, authService *AuthorizationService, db *sql.DB, cache pkgcache.Cache, mdRenderer *markdown.Renderer, mentionService *MentionService, referenceService *IssueReferenceService) *IssueService {
	return &IssueService{
		issueRepo:        issueRepo,
		watcherRepo:      watcherRepo,
		authService:      authService,
		db:               db,
		cache:            cache,
		markdownRenderer: mdRenderer,
		mentionService:   mentionService,
		referenceService: referenceService,
	}
}

// Create creates a new issue
func (s *IssueService) Create(ctx context.Context, projectID int, req *models.CreateIssueRequest, userID int) (*models.Issue, error) {
	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Create issue
	priority := models.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	issue := &models.Issue{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.IssueStatusOpen,
		Priority:    priority,
		AssigneeID:  req.AssigneeID,
		ReporterID:  userID,
		ColumnID:    req.ColumnID,
		MilestoneID: req.MilestoneID,
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

	// TODO: Add labels if provided in req.LabelIDs

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

	// Return updated issue
	return s.issueRepo.GetByID(ctx, id)
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

	err = s.issueRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, issue.ProjectID)

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

	// Return updated issue
	return s.issueRepo.GetByID(ctx, id)
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
