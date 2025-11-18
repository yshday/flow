package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
)

var milestoneTestCounter = 0

func setupMilestoneRepo(t *testing.T) (*MilestoneRepository, *ProjectRepository, *UserRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM milestones")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	milestoneRepo := NewMilestoneRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM milestones")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return milestoneRepo, projectRepo, userRepo, cleanup
}

func createTestProjectForMilestone(t *testing.T, projectRepo *ProjectRepository, userRepo *UserRepository) (*models.Project, *models.User) {
	// Use counter to create unique identifiers
	milestoneTestCounter++
	suffix := fmt.Sprintf("%d", milestoneTestCounter)

	user := &models.User{
		Email:        "milestone_owner_" + suffix + "@example.com",
		Username:     "milestoneowner_" + suffix,
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	project := &models.Project{
		Name:        "Milestone Test Project " + suffix,
		Key:         "MTP" + suffix,
		Description: stringPtr("Test project for milestones"),
		OwnerID:     createdUser.ID,
	}

	createdProject, err := projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return createdProject, createdUser
}

func TestMilestoneRepository_Create(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMilestone(t, projectRepo, userRepo)

	dueDate := time.Now().Add(30 * 24 * time.Hour)
	milestone := &models.Milestone{
		ProjectID:   project.ID,
		Title:       "Version 1.0",
		Description: stringPtr("First major release"),
		DueDate:     &dueDate,
		Status:      models.MilestoneStatusOpen,
	}

	created, err := milestoneRepo.Create(ctx, milestone)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}

	if created.Title != "Version 1.0" {
		t.Errorf("Expected title 'Version 1.0', got '%s'", created.Title)
	}

	if created.Status != models.MilestoneStatusOpen {
		t.Errorf("Expected status 'open', got '%s'", created.Status)
	}

	if created.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if created.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestMilestoneRepository_GetByID(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMilestone(t, projectRepo, userRepo)

	milestone := &models.Milestone{
		ProjectID: project.ID,
		Title:     "Test Milestone",
		Status:    models.MilestoneStatusOpen,
	}

	created, err := milestoneRepo.Create(ctx, milestone)
	if err != nil {
		t.Fatalf("Failed to create milestone: %v", err)
	}

	// Get by ID
	retrieved, err := milestoneRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.Title != "Test Milestone" {
		t.Errorf("Expected title 'Test Milestone', got '%s'", retrieved.Title)
	}
}

func TestMilestoneRepository_GetByID_NotFound(t *testing.T) {
	milestoneRepo, _, _, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()

	_, err := milestoneRepo.GetByID(ctx, 99999)
	if err == nil {
		t.Error("Expected error for non-existent milestone, got nil")
	}
}

func TestMilestoneRepository_ListByProjectID(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMilestone(t, projectRepo, userRepo)

	// Create multiple milestones
	for i := 1; i <= 3; i++ {
		milestone := &models.Milestone{
			ProjectID: project.ID,
			Title:     "Milestone " + string(rune('0'+i)),
			Status:    models.MilestoneStatusOpen,
		}

		_, err := milestoneRepo.Create(ctx, milestone)
		if err != nil {
			t.Fatalf("Failed to create milestone %d: %v", i, err)
		}
	}

	// List milestones
	milestones, err := milestoneRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(milestones) != 3 {
		t.Errorf("Expected 3 milestones, got %d", len(milestones))
	}
}

func TestMilestoneRepository_Update(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMilestone(t, projectRepo, userRepo)

	milestone := &models.Milestone{
		ProjectID: project.ID,
		Title:     "Original Title",
		Status:    models.MilestoneStatusOpen,
	}

	created, err := milestoneRepo.Create(ctx, milestone)
	if err != nil {
		t.Fatalf("Failed to create milestone: %v", err)
	}

	// Update
	created.Title = "Updated Title"
	created.Status = models.MilestoneStatusClosed

	updated, err := milestoneRepo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}

	if updated.Status != models.MilestoneStatusClosed {
		t.Errorf("Expected status 'closed', got '%s'", updated.Status)
	}

	// Verify updated_at changed
	if !updated.UpdatedAt.After(created.CreatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestMilestoneRepository_Delete(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForMilestone(t, projectRepo, userRepo)

	milestone := &models.Milestone{
		ProjectID: project.ID,
		Title:     "To Be Deleted",
		Status:    models.MilestoneStatusOpen,
	}

	created, err := milestoneRepo.Create(ctx, milestone)
	if err != nil {
		t.Fatalf("Failed to create milestone: %v", err)
	}

	// Delete
	err = milestoneRepo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = milestoneRepo.GetByID(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after deleting milestone, got nil")
	}
}

func TestMilestoneRepository_GetWithProgress(t *testing.T) {
	milestoneRepo, projectRepo, userRepo, cleanup := setupMilestoneRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForMilestone(t, projectRepo, userRepo)

	// Create milestone
	milestone := &models.Milestone{
		ProjectID: project.ID,
		Title:     "Test Progress",
		Status:    models.MilestoneStatusOpen,
	}

	created, err := milestoneRepo.Create(ctx, milestone)
	if err != nil {
		t.Fatalf("Failed to create milestone: %v", err)
	}

	// Create board columns for issues
	boardRepo := NewBoardRepository(setupTestDB(t))
	err = boardRepo.CreateDefaultColumns(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to create default columns: %v", err)
	}

	columns, err := boardRepo.ListByProjectID(ctx, project.ID)
	if err != nil || len(columns) == 0 {
		t.Fatalf("Failed to get board columns: %v", err)
	}

	// Create issues associated with milestone
	issueRepo := NewIssueRepository(setupTestDB(t))

	// Create 2 closed issues
	for i := 0; i < 2; i++ {
		issue := &models.Issue{
			ProjectID:   project.ID,
			Title:       "Closed Issue",
			ReporterID:  user.ID,
			ColumnID:    &columns[0].ID,
			Priority:    models.PriorityMedium,
			Status:      models.IssueStatusClosed,
			MilestoneID: &created.ID,
		}
		_, err := issueRepo.Create(ctx, issue)
		if err != nil {
			t.Fatalf("Failed to create closed issue: %v", err)
		}
	}

	// Create 3 open issues
	for i := 0; i < 3; i++ {
		issue := &models.Issue{
			ProjectID:   project.ID,
			Title:       "Open Issue",
			ReporterID:  user.ID,
			ColumnID:    &columns[0].ID,
			Priority:    models.PriorityMedium,
			Status:      models.IssueStatusOpen,
			MilestoneID: &created.ID,
		}
		_, err := issueRepo.Create(ctx, issue)
		if err != nil {
			t.Fatalf("Failed to create open issue: %v", err)
		}
	}

	// Get milestone with progress
	withProgress, err := milestoneRepo.GetWithProgress(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetWithProgress failed: %v", err)
	}

	if withProgress.TotalIssues != 5 {
		t.Errorf("Expected 5 total issues, got %d", withProgress.TotalIssues)
	}

	if withProgress.ClosedIssues != 2 {
		t.Errorf("Expected 2 closed issues, got %d", withProgress.ClosedIssues)
	}

	// Progress should be 40% (2/5 * 100)
	expectedProgress := 40
	if withProgress.Progress != expectedProgress {
		t.Errorf("Expected progress %d%%, got %d%%", expectedProgress, withProgress.Progress)
	}
}
