package service

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupIssueService(t *testing.T) (*IssueService, *repository.UserRepository, *repository.ProjectRepository, *repository.BoardRepository, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	issueRepo := repository.NewIssueRepository(db)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	issueService := NewIssueService(issueRepo, db)

	cleanup := func() {
		db.Exec("DELETE FROM issues WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISVC%')")
		db.Exec("DELETE FROM board_columns WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISVC%')")
		db.Exec("DELETE FROM project_members WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISVC%')")
		db.Exec("DELETE FROM projects WHERE key LIKE 'ISVC%'")
		db.Exec("DELETE FROM users WHERE email LIKE 'issuesvc%@example.com'")
		db.Close()
	}

	return issueService, userRepo, projectRepo, boardRepo, cleanup
}

func TestIssueService_Create(t *testing.T) {
	service, userRepo, projectRepo, boardRepo, cleanup := setupIssueService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	user, _ := userRepo.Create(ctx, &models.User{
		Email:        "issuesvc1@example.com",
		Username:     "issuesvc1",
		PasswordHash: "hash",
	})

	// Create test project
	project, _ := projectRepo.Create(ctx, &models.Project{
		Name:    "Test Project",
		Key:     "ISVC1",
		OwnerID: user.ID,
	})

	// Add user as project member
	db := service.db
	db.ExecContext(ctx, "INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)",
		project.ID, user.ID, models.RoleMember)

	// Create board columns
	boardRepo.CreateDefaultColumns(ctx, project.ID)

	t.Run("should create issue with auto number", func(t *testing.T) {
		desc := "Test description"
		req := &models.CreateIssueRequest{
			Title:       "Test Issue",
			Description: &desc,
		}

		issue, err := service.Create(ctx, project.ID, req, user.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if issue.IssueNumber == 0 {
			t.Error("Expected issue number to be set")
		}

		if issue.Title != req.Title {
			t.Errorf("Expected title %s, got %s", req.Title, issue.Title)
		}

		if issue.ReporterID != user.ID {
			t.Errorf("Expected reporter ID %d, got %d", user.ID, issue.ReporterID)
		}
	})

	t.Run("should deny access if user not project member", func(t *testing.T) {
		// Create another user not in project
		otherUser, _ := userRepo.Create(ctx, &models.User{
			Email:        "issuesvc2@example.com",
			Username:     "issuesvc2",
			PasswordHash: "hash",
		})

		req := &models.CreateIssueRequest{
			Title: "Should Fail",
		}

		_, err := service.Create(ctx, project.ID, req, otherUser.ID)
		if err != pkgerrors.ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})
}

func TestIssueService_GetByID(t *testing.T) {
	service, userRepo, projectRepo, _, cleanup := setupIssueService(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuesvc3@example.com", Username: "issuesvc3", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISVC2", OwnerID: user.ID})
	service.db.ExecContext(ctx, "INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)", project.ID, user.ID, models.RoleMember)

	req := &models.CreateIssueRequest{Title: "Test"}
	issue, _ := service.Create(ctx, project.ID, req, user.ID)

	t.Run("should get issue by ID", func(t *testing.T) {
		found, err := service.GetByID(ctx, issue.ID, user.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != issue.ID {
			t.Errorf("Expected ID %d, got %d", issue.ID, found.ID)
		}
	})

	t.Run("should deny access if user not member", func(t *testing.T) {
		otherUser, _ := userRepo.Create(ctx, &models.User{Email: "issuesvc4@example.com", Username: "issuesvc4", PasswordHash: "hash"})

		_, err := service.GetByID(ctx, issue.ID, otherUser.ID)
		if err != pkgerrors.ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})
}

func TestIssueService_List(t *testing.T) {
	service, userRepo, projectRepo, _, cleanup := setupIssueService(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuesvc5@example.com", Username: "issuesvc5", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISVC3", OwnerID: user.ID})
	service.db.ExecContext(ctx, "INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)", project.ID, user.ID, models.RoleMember)

	// Create issues
	service.Create(ctx, project.ID, &models.CreateIssueRequest{Title: "Issue 1"}, user.ID)
	service.Create(ctx, project.ID, &models.CreateIssueRequest{Title: "Issue 2"}, user.ID)

	t.Run("should list issues for project", func(t *testing.T) {
		filter := &models.IssueFilter{
			ProjectID: project.ID,
		}

		issues, err := service.List(ctx, filter, user.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}
	})
}

func TestIssueService_Update(t *testing.T) {
	service, userRepo, projectRepo, _, cleanup := setupIssueService(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuesvc6@example.com", Username: "issuesvc6", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISVC4", OwnerID: user.ID})
	service.db.ExecContext(ctx, "INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)", project.ID, user.ID, models.RoleMember)

	issue, _ := service.Create(ctx, project.ID, &models.CreateIssueRequest{Title: "Original"}, user.ID)

	t.Run("should update issue", func(t *testing.T) {
		newTitle := "Updated"
		closedStatus := models.IssueStatusClosed
		req := &models.UpdateIssueRequest{
			Title:  &newTitle,
			Status: &closedStatus,
		}

		updated, err := service.Update(ctx, issue.ID, req, user.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updated.Title != newTitle {
			t.Errorf("Expected title %s, got %s", newTitle, updated.Title)
		}

		if updated.Status != closedStatus {
			t.Errorf("Expected status closed, got %s", updated.Status)
		}
	})
}

func TestIssueService_Delete(t *testing.T) {
	service, userRepo, projectRepo, _, cleanup := setupIssueService(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuesvc7@example.com", Username: "issuesvc7", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISVC5", OwnerID: user.ID})
	service.db.ExecContext(ctx, "INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)", project.ID, user.ID, models.RoleOwner)

	issue, _ := service.Create(ctx, project.ID, &models.CreateIssueRequest{Title: "To Delete"}, user.ID)

	t.Run("should allow owner to delete", func(t *testing.T) {
		err := service.Delete(ctx, issue.ID, user.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify deletion
		_, err = service.GetByID(ctx, issue.ID, user.ID)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}
