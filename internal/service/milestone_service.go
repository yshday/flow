package service

import (
	"context"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
)

// MilestoneService handles milestone business logic
type MilestoneService struct {
	milestoneRepo *repository.MilestoneRepository
	projectRepo   *repository.ProjectRepository
	authService   *AuthorizationService
	cache         pkgcache.Cache
}

// NewMilestoneService creates a new milestone service
func NewMilestoneService(
	milestoneRepo *repository.MilestoneRepository,
	projectRepo *repository.ProjectRepository,
	authService *AuthorizationService,
	cache pkgcache.Cache,
) *MilestoneService {
	return &MilestoneService{
		milestoneRepo: milestoneRepo,
		projectRepo:   projectRepo,
		authService:   authService,
		cache:         cache,
	}
}

// Create creates a new milestone
func (s *MilestoneService) Create(ctx context.Context, projectID int, req *models.CreateMilestoneRequest, userID int) (*models.Milestone, error) {
	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	milestone := &models.Milestone{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      models.MilestoneStatusOpen,
	}

	created, err := s.milestoneRepo.Create(ctx, milestone)
	if err != nil {
		return nil, err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, projectID)

	return created, nil
}

// GetByID retrieves a milestone by ID
func (s *MilestoneService) GetByID(ctx context.Context, id int, userID int) (*models.Milestone, error) {
	milestone, err := s.milestoneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, milestone.ProjectID, userID); err != nil {
		return nil, err
	}

	return milestone, nil
}

// ListByProjectID lists all milestones for a project
func (s *MilestoneService) ListByProjectID(ctx context.Context, projectID int, userID int) ([]*models.Milestone, error) {
	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	return s.milestoneRepo.ListByProjectID(ctx, projectID)
}

// Update updates a milestone
func (s *MilestoneService) Update(ctx context.Context, id int, req *models.UpdateMilestoneRequest, userID int) (*models.Milestone, error) {
	// Get existing milestone
	milestone, err := s.milestoneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, milestone.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		milestone.Title = *req.Title
	}
	if req.Description != nil {
		milestone.Description = req.Description
	}
	if req.DueDate != nil {
		milestone.DueDate = req.DueDate
	}
	if req.Status != nil {
		milestone.Status = *req.Status
	}

	updated, err := s.milestoneRepo.Update(ctx, milestone)
	if err != nil {
		return nil, err
	}

	// Invalidate project caches (milestone status affects project stats)
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, milestone.ProjectID)

	return updated, nil
}

// Delete deletes a milestone
func (s *MilestoneService) Delete(ctx context.Context, id int, userID int) error {
	// Get existing milestone
	milestone, err := s.milestoneRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user has admin permission (only admins/owners can delete milestones)
	if err := s.authService.CheckAdminPermission(ctx, milestone.ProjectID, userID); err != nil {
		return err
	}

	err = s.milestoneRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, milestone.ProjectID)

	return nil
}

// GetWithProgress retrieves a milestone with progress calculation
func (s *MilestoneService) GetWithProgress(ctx context.Context, id int, userID int) (*models.Milestone, error) {
	milestone, err := s.milestoneRepo.GetWithProgress(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, milestone.ProjectID, userID); err != nil {
		return nil, err
	}

	return milestone, nil
}
