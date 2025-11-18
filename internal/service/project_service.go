package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgcache "github.com/yourusername/issue-tracker/pkg/cache"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repository.ProjectRepository
	boardRepo   *repository.BoardRepository
	db          *sql.DB
	cache       pkgcache.Cache
}

// NewProjectService creates a new project service
func NewProjectService(
	projectRepo *repository.ProjectRepository,
	boardRepo *repository.BoardRepository,
	db *sql.DB,
	cache pkgcache.Cache,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		boardRepo:   boardRepo,
		db:          db,
		cache:       cache,
	}
}

// Create creates a new project with default columns and adds owner as member
func (s *ProjectService) Create(ctx context.Context, req *models.CreateProjectRequest, ownerID int) (*models.Project, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create project
	project := &models.Project{
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	createdProject, err := s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	// Create default board columns
	err = s.boardRepo.CreateDefaultColumns(ctx, createdProject.ID)
	if err != nil {
		return nil, err
	}

	// Add owner as member with owner role
	_, err = tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
	`, createdProject.ID, ownerID, models.RoleOwner)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return createdProject, nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id int, userID int) (*models.Project, error) {
	// First check if project exists
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then check if user has access to the project
	hasAccess, err := s.userHasAccess(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return project, nil
}

// List retrieves all projects accessible by a user
func (s *ProjectService) List(ctx context.Context, userID int) ([]*models.Project, error) {
	return s.projectRepo.ListByUserID(ctx, userID)
}

// Update updates a project
func (s *ProjectService) Update(ctx context.Context, id int, req *models.UpdateProjectRequest, userID int) (*models.Project, error) {
	// Check if user has admin or owner role
	hasPermission, err := s.userHasPermission(ctx, userID, id, []models.ProjectRole{models.RoleOwner, models.RoleAdmin})
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, pkgerrors.ErrForbidden
	}

	// Get existing project
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}

	// Save changes
	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, id)

	return project, nil
}

// Delete deletes a project
func (s *ProjectService) Delete(ctx context.Context, id int, userID int) error {
	// Only owner can delete project
	hasPermission, err := s.userHasPermission(ctx, userID, id, []models.ProjectRole{models.RoleOwner})
	if err != nil {
		return err
	}

	if !hasPermission {
		return pkgerrors.ErrForbidden
	}

	err = s.projectRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate project caches
	_ = pkgcache.InvalidateAllProjectCaches(ctx, s.cache, id)

	return nil
}

// userHasAccess checks if user has any access to the project
func (s *ProjectService) userHasAccess(ctx context.Context, userID int, projectID int) (bool, error) {
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
func (s *ProjectService) userHasPermission(ctx context.Context, userID int, projectID int, allowedRoles []models.ProjectRole) (bool, error) {
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
