package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ProjectMemberService handles project member business logic
type ProjectMemberService struct {
	memberRepo  *repository.ProjectMemberRepository
	projectRepo *repository.ProjectRepository
	userRepo    *repository.UserRepository
	db          *sql.DB
}

// NewProjectMemberService creates a new project member service
func NewProjectMemberService(
	memberRepo *repository.ProjectMemberRepository,
	projectRepo *repository.ProjectRepository,
	userRepo *repository.UserRepository,
	db *sql.DB,
) *ProjectMemberService {
	return &ProjectMemberService{
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		db:          db,
	}
}

// AddMember adds a new member to a project
func (s *ProjectMemberService) AddMember(ctx context.Context, projectID int, req *models.AddMemberRequest, currentUserID int) error {
	// Check if project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Check if current user has permission (owner or admin)
	hasPermission, err := s.userHasPermission(ctx, currentUserID, projectID, []models.ProjectRole{models.RoleOwner, models.RoleAdmin})
	if err != nil {
		return err
	}

	if !hasPermission {
		return pkgerrors.ErrForbidden
	}

	// Check if user to be added exists
	_, err = s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// Add member
	member := &models.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      string(req.Role),
		InvitedBy: &currentUserID,
	}

	return s.memberRepo.AddMember(ctx, member)
}

// ListMembers lists all members of a project
func (s *ProjectMemberService) ListMembers(ctx context.Context, projectID int, currentUserID int) ([]*models.ProjectMember, error) {
	// Check if user has access to project
	hasAccess, err := s.userHasAccess(ctx, currentUserID, projectID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, pkgerrors.ErrForbidden
	}

	return s.memberRepo.ListByProjectID(ctx, projectID)
}

// UpdateMemberRole updates a member's role
func (s *ProjectMemberService) UpdateMemberRole(ctx context.Context, projectID, memberUserID int, req *models.UpdateMemberRoleRequest, currentUserID int) error {
	// Check if current user has permission (owner or admin)
	hasPermission, err := s.userHasPermission(ctx, currentUserID, projectID, []models.ProjectRole{models.RoleOwner, models.RoleAdmin})
	if err != nil {
		return err
	}

	if !hasPermission {
		return pkgerrors.ErrForbidden
	}

	// Check if member exists
	_, err = s.memberRepo.GetMember(ctx, projectID, memberUserID)
	if err != nil {
		return err
	}

	// Update role
	return s.memberRepo.UpdateRole(ctx, projectID, memberUserID, string(req.Role))
}

// RemoveMember removes a member from a project
func (s *ProjectMemberService) RemoveMember(ctx context.Context, projectID, memberUserID int, currentUserID int) error {
	// Check if current user has permission (owner or admin)
	hasPermission, err := s.userHasPermission(ctx, currentUserID, projectID, []models.ProjectRole{models.RoleOwner, models.RoleAdmin})
	if err != nil {
		return err
	}

	if !hasPermission {
		return pkgerrors.ErrForbidden
	}

	// Check if member exists
	_, err = s.memberRepo.GetMember(ctx, projectID, memberUserID)
	if err != nil {
		return err
	}

	// Cannot remove the owner
	member, _ := s.memberRepo.GetMember(ctx, projectID, memberUserID)
	if member != nil && member.Role == string(models.RoleOwner) {
		return pkgerrors.ErrValidation
	}

	// Remove member
	return s.memberRepo.RemoveMember(ctx, projectID, memberUserID)
}

// userHasAccess checks if user has any access to the project
func (s *ProjectMemberService) userHasAccess(ctx context.Context, userID int, projectID int) (bool, error) {
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
func (s *ProjectMemberService) userHasPermission(ctx context.Context, userID int, projectID int, allowedRoles []models.ProjectRole) (bool, error) {
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
