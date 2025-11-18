package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ActivityService handles activity business logic
type ActivityService struct {
	activityRepo *repository.ActivityRepository
	projectRepo  *repository.ProjectRepository
	issueRepo    *repository.IssueRepository
	db           *sql.DB
}

// NewActivityService creates a new activity service
func NewActivityService(
	activityRepo *repository.ActivityRepository,
	projectRepo *repository.ProjectRepository,
	issueRepo *repository.IssueRepository,
	db *sql.DB,
) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		projectRepo:  projectRepo,
		issueRepo:    issueRepo,
		db:           db,
	}
}

// LogActivity creates a new activity log entry
func (s *ActivityService) LogActivity(ctx context.Context, req *models.CreateActivityRequest) (*models.Activity, error) {
	activity := &models.Activity{
		ProjectID:  req.ProjectID,
		IssueID:    req.IssueID,
		UserID:     req.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		FieldName:  req.FieldName,
		OldValue:   req.OldValue,
		NewValue:   req.NewValue,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		Metadata:   req.Metadata,
	}

	return s.activityRepo.Create(ctx, activity)
}

// ListByProjectID lists all activities for a project
func (s *ActivityService) ListByProjectID(ctx context.Context, projectID int, limit, offset int, userID int) ([]*models.Activity, error) {
	// Check if user has access to project
	hasAccess, err := s.userHasAccessToProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.activityRepo.ListByProjectID(ctx, projectID, limit, offset)
}

// ListByIssueID lists all activities for an issue
func (s *ActivityService) ListByIssueID(ctx context.Context, issueID int, limit, offset int, userID int) ([]*models.Activity, error) {
	// Get issue to check project access
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project
	hasAccess, err := s.userHasAccessToProject(ctx, userID, issue.ProjectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.activityRepo.ListByIssueID(ctx, issueID, limit, offset)
}

// ListWithFilter lists activities with filtering
func (s *ActivityService) ListWithFilter(ctx context.Context, filter *models.ActivityFilter, userID int) ([]*models.Activity, error) {
	// If project filter is specified, check access
	if filter.ProjectID != nil {
		hasAccess, err := s.userHasAccessToProject(ctx, userID, *filter.ProjectID)
		if err != nil {
			return nil, err
		}

		if !hasAccess {
			return nil, pkgerrors.ErrForbidden
		}
	}

	// If issue filter is specified, check access
	if filter.IssueID != nil {
		issue, err := s.issueRepo.GetByID(ctx, *filter.IssueID)
		if err != nil {
			return nil, err
		}

		hasAccess, err := s.userHasAccessToProject(ctx, userID, issue.ProjectID)
		if err != nil {
			return nil, err
		}

		if !hasAccess {
			return nil, pkgerrors.ErrForbidden
		}
	}

	return s.activityRepo.List(ctx, filter)
}

// userHasAccessToProject checks if user has access to a project
func (s *ActivityService) userHasAccessToProject(ctx context.Context, userID int, projectID int) (bool, error) {
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
