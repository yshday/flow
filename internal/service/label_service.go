package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// LabelService handles label business logic
type LabelService struct {
	labelRepo   *repository.LabelRepository
	projectRepo *repository.ProjectRepository
	issueRepo   *repository.IssueRepository
	authService *AuthorizationService
	db          *sql.DB
	cache       pkgcache.Cache
}

// NewLabelService creates a new label service
func NewLabelService(labelRepo *repository.LabelRepository, projectRepo *repository.ProjectRepository, issueRepo *repository.IssueRepository, authService *AuthorizationService, db *sql.DB, cache pkgcache.Cache) *LabelService {
	return &LabelService{
		labelRepo:   labelRepo,
		projectRepo: projectRepo,
		issueRepo:   issueRepo,
		authService: authService,
		db:          db,
		cache:       cache,
	}
}

// Create creates a new label
func (s *LabelService) Create(ctx context.Context, projectID int, req *models.CreateLabelRequest, userID int) (*models.Label, error) {
	// Check if user has admin permission (only admins/owners can create labels)
	if err := s.authService.CheckAdminPermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	label := &models.Label{
		ProjectID: projectID,
		Name:      req.Name,
		Color:     req.Color,
	}

	created, err := s.labelRepo.Create(ctx, label)
	if err != nil {
		return nil, err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, projectID)

	return created, nil
}

// GetByID retrieves a label by ID
func (s *LabelService) GetByID(ctx context.Context, id int, userID int) (*models.Label, error) {
	label, err := s.labelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, label.ProjectID, userID); err != nil {
		return nil, err
	}

	return label, nil
}

// List lists all labels for a project
func (s *LabelService) List(ctx context.Context, projectID int, userID int) ([]*models.Label, error) {
	// Check if user has access to project
	if err := s.authService.CheckProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	return s.labelRepo.ListByProjectID(ctx, projectID)
}

// Update updates a label
func (s *LabelService) Update(ctx context.Context, id int, req *models.UpdateLabelRequest, userID int) (*models.Label, error) {
	// Get existing label
	label, err := s.labelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has admin permission (only admins/owners can update labels)
	if err := s.authService.CheckAdminPermission(ctx, label.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		label.Name = *req.Name
	}
	if req.Color != nil {
		label.Color = *req.Color
	}

	err = s.labelRepo.Update(ctx, label)
	if err != nil {
		return nil, err
	}

	return s.labelRepo.GetByID(ctx, id)
}

// Delete deletes a label
func (s *LabelService) Delete(ctx context.Context, id int, userID int) error {
	// Get label
	label, err := s.labelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has admin permission (only admins/owners can delete labels)
	if err := s.authService.CheckAdminPermission(ctx, label.ProjectID, userID); err != nil {
		return err
	}

	err = s.labelRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, label.ProjectID)

	return nil
}

// AddToIssue adds a label to an issue
func (s *LabelService) AddToIssue(ctx context.Context, issueID int, labelID int, userID int) error {
	// Get issue to check it exists and get project ID
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return err
	}

	// Get label to check it exists and validate same project
	label, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		return err
	}

	// Verify label belongs to same project as issue
	if label.ProjectID != issue.ProjectID {
		return pkgerrors.ErrValidation
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	return s.labelRepo.AddToIssue(ctx, issueID, labelID)
}

// RemoveFromIssue removes a label from an issue
func (s *LabelService) RemoveFromIssue(ctx context.Context, issueID int, labelID int, userID int) error {
	// Get issue to check it exists and get project ID
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return err
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	return s.labelRepo.RemoveFromIssue(ctx, issueID, labelID)
}

// ListByIssueID lists all labels for an issue
func (s *LabelService) ListByIssueID(ctx context.Context, issueID int, userID int) ([]*models.Label, error) {
	// Get issue to check it exists and get project ID
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return s.labelRepo.ListByIssueID(ctx, issueID)
}
