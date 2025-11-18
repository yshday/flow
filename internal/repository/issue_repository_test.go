package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupIssueRepo(t *testing.T) (*IssueRepository, *UserRepository, *ProjectRepository, *BoardRepository, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	issueRepo := NewIssueRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)
	boardRepo := NewBoardRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM issues WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISS%')")
		db.Exec("DELETE FROM board_columns WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISS%')")
		db.Exec("DELETE FROM project_members WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'ISS%')")
		db.Exec("DELETE FROM projects WHERE key LIKE 'ISS%'")
		db.Exec("DELETE FROM users WHERE email LIKE 'issuetest%@example.com'")
		db.Close()
	}

	return issueRepo, userRepo, projectRepo, boardRepo, cleanup
}

func TestIssueRepository_Create(t *testing.T) {
	issueRepo, userRepo, projectRepo, boardRepo, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	user := &models.User{
		Email:        "issuetest1@example.com",
		Username:     "issuetest1",
		PasswordHash: "hash",
	}
	createdUser, _ := userRepo.Create(ctx, user)

	// Create test project
	project := &models.Project{
		Name:    "Issue Test Project",
		Key:     "ISS1",
		OwnerID: createdUser.ID,
	}
	createdProject, _ := projectRepo.Create(ctx, project)

	// Create default columns
	boardRepo.CreateDefaultColumns(ctx, createdProject.ID)

	t.Run("should create issue with auto-generated number", func(t *testing.T) {
		desc := "Test issue description"
		issue := &models.Issue{
			ProjectID:   createdProject.ID,
			Title:       "Test Issue",
			Description: &desc,
			Status:      models.IssueStatusOpen,
			Priority:    models.PriorityMedium,
			ReporterID:  createdUser.ID,
		}

		created, err := issueRepo.Create(ctx, issue)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if created.ID == 0 {
			t.Error("Expected ID to be set")
		}

		if created.IssueNumber != 1 {
			t.Errorf("Expected issue number 1, got %d", created.IssueNumber)
		}

		if created.Title != issue.Title {
			t.Errorf("Expected title %s, got %s", issue.Title, created.Title)
		}

		if created.Version != 1 {
			t.Errorf("Expected version 1, got %d", created.Version)
		}
	})

	t.Run("should auto-increment issue number per project", func(t *testing.T) {
		issue1 := &models.Issue{
			ProjectID:  createdProject.ID,
			Title:      "Issue 2",
			Status:     models.IssueStatusOpen,
			Priority:   models.PriorityMedium,
			ReporterID: createdUser.ID,
		}
		created1, _ := issueRepo.Create(ctx, issue1)

		issue2 := &models.Issue{
			ProjectID:  createdProject.ID,
			Title:      "Issue 3",
			Status:     models.IssueStatusOpen,
			Priority:   models.PriorityMedium,
			ReporterID: createdUser.ID,
		}
		created2, _ := issueRepo.Create(ctx, issue2)

		if created2.IssueNumber != created1.IssueNumber+1 {
			t.Errorf("Expected issue numbers to increment, got %d and %d", created1.IssueNumber, created2.IssueNumber)
		}
	})
}

func TestIssueRepository_GetByID(t *testing.T) {
	issueRepo, userRepo, projectRepo, _, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuetest2@example.com", Username: "issuetest2", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISS2", OwnerID: user.ID})
	issue := &models.Issue{
		ProjectID:  project.ID,
		Title:      "Test Issue",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityMedium,
		ReporterID: user.ID,
	}
	created, _ := issueRepo.Create(ctx, issue)

	t.Run("should get issue by ID", func(t *testing.T) {
		found, err := issueRepo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
		}

		if found.Title != created.Title {
			t.Errorf("Expected title %s, got %s", created.Title, found.Title)
		}
	})

	t.Run("should return ErrNotFound for non-existent ID", func(t *testing.T) {
		_, err := issueRepo.GetByID(ctx, 999999)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestIssueRepository_GetByProjectAndNumber(t *testing.T) {
	issueRepo, userRepo, projectRepo, _, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuetest3@example.com", Username: "issuetest3", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISS3", OwnerID: user.ID})
	issue, _ := issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "Test Issue",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityMedium,
		ReporterID: user.ID,
	})

	t.Run("should get issue by project and number", func(t *testing.T) {
		found, err := issueRepo.GetByProjectAndNumber(ctx, project.ID, issue.IssueNumber)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != issue.ID {
			t.Errorf("Expected ID %d, got %d", issue.ID, found.ID)
		}
	})

	t.Run("should return ErrNotFound for non-existent number", func(t *testing.T) {
		_, err := issueRepo.GetByProjectAndNumber(ctx, project.ID, 999)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestIssueRepository_List(t *testing.T) {
	issueRepo, userRepo, projectRepo, _, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user1, _ := userRepo.Create(ctx, &models.User{Email: "issuetest4@example.com", Username: "issuetest4", PasswordHash: "hash"})
	user2, _ := userRepo.Create(ctx, &models.User{Email: "issuetest5@example.com", Username: "issuetest5", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISS4", OwnerID: user1.ID})

	// Create issues with different attributes
	issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "Bug in login",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityHigh,
		ReporterID: user1.ID,
		AssigneeID: &user2.ID,
	})

	issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "Feature request",
		Status:     models.IssueStatusClosed,
		Priority:   models.PriorityLow,
		ReporterID: user1.ID,
	})

	issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "Documentation update",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityMedium,
		ReporterID: user2.ID,
	})

	t.Run("should list all issues for project", func(t *testing.T) {
		filter := &models.IssueFilter{
			ProjectID: project.ID,
		}

		issues, err := issueRepo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 3 {
			t.Errorf("Expected 3 issues, got %d", len(issues))
		}
	})

	t.Run("should filter by status", func(t *testing.T) {
		openStatus := models.IssueStatusOpen
		filter := &models.IssueFilter{
			ProjectID: project.ID,
			Status:    &openStatus,
		}

		issues, err := issueRepo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 2 {
			t.Errorf("Expected 2 open issues, got %d", len(issues))
		}
	})

	t.Run("should filter by in_progress status", func(t *testing.T) {
		// Create an issue with in_progress status
		issueRepo.Create(ctx, &models.Issue{
			ProjectID:  project.ID,
			Title:      "Work in progress",
			Status:     models.IssueStatusInProgress,
			Priority:   models.PriorityMedium,
			ReporterID: user1.ID,
		})

		inProgressStatus := models.IssueStatusInProgress
		filter := &models.IssueFilter{
			ProjectID: project.ID,
			Status:    &inProgressStatus,
		}

		issues, err := issueRepo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 1 {
			t.Errorf("Expected 1 in_progress issue, got %d", len(issues))
		}

		if issues[0].Status != models.IssueStatusInProgress {
			t.Errorf("Expected status in_progress, got %s", issues[0].Status)
		}
	})

	t.Run("should filter by assignee", func(t *testing.T) {
		filter := &models.IssueFilter{
			ProjectID:  project.ID,
			AssigneeID: &user2.ID,
		}

		issues, err := issueRepo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 1 {
			t.Errorf("Expected 1 issue assigned to user2, got %d", len(issues))
		}
	})

	t.Run("should search by title", func(t *testing.T) {
		filter := &models.IssueFilter{
			ProjectID: project.ID,
			Search:    "login",
		}

		issues, err := issueRepo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(issues) != 1 {
			t.Errorf("Expected 1 issue with 'login' in title, got %d", len(issues))
		}
	})
}

func TestIssueRepository_Update(t *testing.T) {
	issueRepo, userRepo, projectRepo, _, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuetest6@example.com", Username: "issuetest6", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISS5", OwnerID: user.ID})
	issue, _ := issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "Original Title",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityMedium,
		ReporterID: user.ID,
	})

	t.Run("should update issue", func(t *testing.T) {
		issue.Title = "Updated Title"
		issue.Status = models.IssueStatusClosed
		issue.Priority = models.PriorityHigh

		err := issueRepo.Update(ctx, issue)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify update
		updated, _ := issueRepo.GetByID(ctx, issue.ID)
		if updated.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got %s", updated.Title)
		}

		if updated.Status != models.IssueStatusClosed {
			t.Errorf("Expected status closed, got %s", updated.Status)
		}

		if updated.Version != 2 {
			t.Errorf("Expected version 2, got %d", updated.Version)
		}
	})

	t.Run("should fail update with wrong version (optimistic locking)", func(t *testing.T) {
		// Get current issue
		current, _ := issueRepo.GetByID(ctx, issue.ID)

		// Update with old version
		current.Version = 1
		current.Title = "Should Fail"

		err := issueRepo.Update(ctx, current)
		if err != pkgerrors.ErrConflict {
			t.Errorf("Expected ErrConflict, got %v", err)
		}
	})
}

func TestIssueRepository_Delete(t *testing.T) {
	issueRepo, userRepo, projectRepo, _, cleanup := setupIssueRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	user, _ := userRepo.Create(ctx, &models.User{Email: "issuetest7@example.com", Username: "issuetest7", PasswordHash: "hash"})
	project, _ := projectRepo.Create(ctx, &models.Project{Name: "Test", Key: "ISS6", OwnerID: user.ID})
	issue, _ := issueRepo.Create(ctx, &models.Issue{
		ProjectID:  project.ID,
		Title:      "To Delete",
		Status:     models.IssueStatusOpen,
		Priority:   models.PriorityMedium,
		ReporterID: user.ID,
	})

	t.Run("should soft delete issue", func(t *testing.T) {
		err := issueRepo.Delete(ctx, issue.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify soft deletion - should return ErrNotFound
		_, err = issueRepo.GetByID(ctx, issue.ID)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound for deleted issue, got %v", err)
		}
	})
}
