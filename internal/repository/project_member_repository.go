package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// ProjectMemberRepository handles project member data operations
type ProjectMemberRepository struct {
	db *sql.DB
}

// NewProjectMemberRepository creates a new project member repository
func NewProjectMemberRepository(db *sql.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{
		db: db,
	}
}

// AddMember adds a new member to a project
func (r *ProjectMemberRepository) AddMember(ctx context.Context, member *models.ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, member.ProjectID, member.UserID, member.Role, member.InvitedBy)
	return err
}

// ListByProjectID retrieves all members of a project with their user information
func (r *ProjectMemberRepository) ListByProjectID(ctx context.Context, projectID int) ([]*models.ProjectMember, error) {
	query := `
		SELECT pm.project_id, pm.user_id, pm.role, pm.joined_at, pm.invited_by,
		       u.id, u.email, u.username, u.created_at, u.updated_at
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.ProjectMember, 0)
	for rows.Next() {
		member := &models.ProjectMember{
			User: &models.User{},
		}

		err := rows.Scan(
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.InvitedBy,
			&member.User.ID,
			&member.User.Email,
			&member.User.Username,
			&member.User.CreatedAt,
			&member.User.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// GetMember retrieves a specific member of a project
func (r *ProjectMemberRepository) GetMember(ctx context.Context, projectID, userID int) (*models.ProjectMember, error) {
	query := `
		SELECT project_id, user_id, role, joined_at, invited_by
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	member := &models.ProjectMember{}
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(
		&member.ProjectID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.InvitedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return member, nil
}

// UpdateRole updates the role of a project member
func (r *ProjectMemberRepository) UpdateRole(ctx context.Context, projectID, userID int, role string) error {
	query := `
		UPDATE project_members
		SET role = $1
		WHERE project_id = $2 AND user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, role, projectID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// RemoveMember removes a member from a project
func (r *ProjectMemberRepository) RemoveMember(ctx context.Context, projectID, userID int) error {
	query := `
		DELETE FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, projectID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// ListByUserID retrieves all projects a user is a member of
func (r *ProjectMemberRepository) ListByUserID(ctx context.Context, userID int) ([]*models.ProjectMember, error) {
	query := `
		SELECT project_id, user_id, role, joined_at, invited_by
		FROM project_members
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.ProjectMember, 0)
	for rows.Next() {
		member := &models.ProjectMember{}

		err := rows.Scan(
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.InvitedBy,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// ListByUserIDWithProjects retrieves all projects a user is a member of with project details
func (r *ProjectMemberRepository) ListByUserIDWithProjects(ctx context.Context, userID int) ([]*models.ProjectMember, error) {
	query := `
		SELECT pm.project_id, pm.user_id, pm.role, pm.joined_at, pm.invited_by,
		       p.id, p.name, p.key, p.description, p.owner_id, p.created_at, p.updated_at
		FROM project_members pm
		JOIN projects p ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY pm.joined_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.ProjectMember, 0)
	for rows.Next() {
		member := &models.ProjectMember{
			Project: &models.Project{},
		}

		var description sql.NullString
		err := rows.Scan(
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.InvitedBy,
			&member.Project.ID,
			&member.Project.Name,
			&member.Project.Key,
			&description,
			&member.Project.OwnerID,
			&member.Project.CreatedAt,
			&member.Project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			member.Project.Description = &description.String
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}
