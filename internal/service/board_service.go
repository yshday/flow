package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
)

// BoardService handles board business logic
type BoardService struct {
	boardRepo   *repository.BoardRepository
	projectRepo *repository.ProjectRepository
	authService *AuthorizationService
	db          *sql.DB
}

// NewBoardService creates a new board service
func NewBoardService(boardRepo *repository.BoardRepository, projectRepo *repository.ProjectRepository, authService *AuthorizationService, db *sql.DB) *BoardService {
	return &BoardService{
		boardRepo:   boardRepo,
		projectRepo: projectRepo,
		authService: authService,
		db:          db,
	}
}

// List lists all board columns for a project
func (s *BoardService) List(ctx context.Context, projectID int, userID int) ([]*models.BoardColumn, error) {
	// Check if user has access to project
	if err := s.authService.CheckProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	return s.boardRepo.ListByProjectID(ctx, projectID)
}

// CreateColumn creates a new board column
func (s *BoardService) CreateColumn(ctx context.Context, projectID int, req *models.CreateBoardColumnRequest, userID int) (*models.BoardColumn, error) {
	// Check if user has admin permission (only admins/owners can create columns)
	if err := s.authService.CheckAdminPermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	column := &models.BoardColumn{
		ProjectID: projectID,
		Name:      req.Name,
		Position:  req.Position,
	}

	return s.boardRepo.Create(ctx, column)
}

// UpdateColumn updates a board column
func (s *BoardService) UpdateColumn(ctx context.Context, columnID int, req *models.UpdateBoardColumnRequest, userID int) (*models.BoardColumn, error) {
	// Get existing column
	column, err := s.boardRepo.GetByID(ctx, columnID)
	if err != nil {
		return nil, err
	}

	// Check if user has admin permission (only admins/owners can update columns)
	if err := s.authService.CheckAdminPermission(ctx, column.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		column.Name = *req.Name
	}
	if req.Position != nil {
		column.Position = *req.Position
	}

	err = s.boardRepo.Update(ctx, column)
	if err != nil {
		return nil, err
	}

	return s.boardRepo.GetByID(ctx, columnID)
}

// DeleteColumn deletes a board column
func (s *BoardService) DeleteColumn(ctx context.Context, columnID int, userID int) error {
	// Get column
	column, err := s.boardRepo.GetByID(ctx, columnID)
	if err != nil {
		return err
	}

	// Check if user has admin permission (only admins/owners can delete columns)
	if err := s.authService.CheckAdminPermission(ctx, column.ProjectID, userID); err != nil {
		return err
	}

	return s.boardRepo.Delete(ctx, columnID)
}
