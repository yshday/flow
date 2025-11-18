package service

import (
	"context"

	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/errors"
)

// AuthorizationService handles common authorization logic
type AuthorizationService struct {
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
) *AuthorizationService {
	return &AuthorizationService{
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

// CheckProjectAccess verifies that a user has access to a project (as owner or member)
// Returns nil if user has access, error otherwise
func (s *AuthorizationService) CheckProjectAccess(ctx context.Context, projectID int, userID int) error {
	// Get project to check ownership
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.NewNotFoundError("project not found")
	}

	// Check if user is the owner
	if project.OwnerID == userID {
		return nil
	}

	// Check if user is a member
	member, err := s.memberRepo.GetMember(ctx, projectID, userID)
	if err != nil || member == nil {
		return errors.NewPermissionError("access denied: you are not a member of this project")
	}

	return nil
}

// CheckWritePermission verifies that a user has write permission in a project
// Viewers have read-only access and cannot perform write operations
// Returns nil if user has write permission, error otherwise
func (s *AuthorizationService) CheckWritePermission(ctx context.Context, projectID int, userID int) error {
	// First check if user has access at all
	err := s.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		return err
	}

	// Get project to check ownership
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.NewNotFoundError("project not found")
	}

	// Owner always has write permission
	if project.OwnerID == userID {
		return nil
	}

	// Check member role
	member, err := s.memberRepo.GetMember(ctx, projectID, userID)
	if err != nil || member == nil {
		return errors.NewPermissionError("access denied: you are not a member of this project")
	}

	// Viewers cannot perform write operations
	if member.Role == "viewer" {
		return errors.NewPermissionError("access denied: viewers have read-only access")
	}

	return nil
}

// CheckAdminPermission verifies that a user has admin or owner permission in a project
// Only owners and admins can perform administrative operations
// Returns nil if user has admin permission, error otherwise
func (s *AuthorizationService) CheckAdminPermission(ctx context.Context, projectID int, userID int) error {
	// First check if user has access at all
	err := s.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		return err
	}

	// Get project to check ownership
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.NewNotFoundError("project not found")
	}

	// Owner always has admin permission
	if project.OwnerID == userID {
		return nil
	}

	// Check member role
	member, err := s.memberRepo.GetMember(ctx, projectID, userID)
	if err != nil || member == nil {
		return errors.NewPermissionError("access denied: you are not a member of this project")
	}

	// Only admins and owners can perform admin operations
	if member.Role != "admin" && member.Role != "owner" {
		return errors.NewPermissionError("access denied: only admins and owners can perform this operation")
	}

	return nil
}

// GetUserRole returns the user's role in a project
// Returns "owner" if user is the owner, otherwise returns the member's role
// Returns error if user is not associated with the project
func (s *AuthorizationService) GetUserRole(ctx context.Context, projectID int, userID int) (string, error) {
	// Get project to check ownership
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return "", errors.NewNotFoundError("project not found")
	}

	// Check if user is the owner
	if project.OwnerID == userID {
		return "owner", nil
	}

	// Get member role
	member, err := s.memberRepo.GetMember(ctx, projectID, userID)
	if err != nil || member == nil {
		return "", errors.NewPermissionError("access denied: you are not a member of this project")
	}

	return member.Role, nil
}
