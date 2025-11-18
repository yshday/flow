package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
)

func setupActivityRepo(t *testing.T) (*ActivityRepository, *UserRepository, *ProjectRepository, *IssueRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM activities")
	db.Exec("DELETE FROM issues")
	db.Exec("DELETE FROM board_columns")
	db.Exec("DELETE FROM project_members")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	activityRepo := NewActivityRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)
	issueRepo := NewIssueRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM activities")
		db.Exec("DELETE FROM issues")
		db.Exec("DELETE FROM board_columns")
		db.Exec("DELETE FROM project_members")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return activityRepo, userRepo, projectRepo, issueRepo, cleanup
}

func createTestUserForActivity(t *testing.T, repo *UserRepository, email string) *models.User {
	user := &models.User{
		Email:        email,
		Username:     email,
		PasswordHash: "hashedpassword",
	}

	created, err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return created
}

func createTestProjectForActivity(t *testing.T, projectRepo *ProjectRepository, userRepo *UserRepository) (*models.Project, *models.User) {
	owner := createTestUserForActivity(t, userRepo, "activity_owner@example.com")

	project := &models.Project{
		Name:        "Activity Test Project",
		Key:         "ATP",
		Description: stringPtr("Test project for activities"),
		OwnerID:     owner.ID,
	}

	created, err := projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return created, owner
}

func TestActivityRepository_Create(t *testing.T) {
	activityRepo, userRepo, projectRepo, _, cleanup := setupActivityRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForActivity(t, projectRepo, userRepo)

	activity := &models.Activity{
		ProjectID:  &project.ID,
		UserID:     user.ID,
		Action:     string(models.ActionCreated),
		EntityType: string(models.EntityTypeProject),
		EntityID:   &project.ID,
	}

	created, err := activityRepo.Create(ctx, activity)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}

	if created.ProjectID == nil || *created.ProjectID != project.ID {
		t.Errorf("Expected ProjectID %d, got %v", project.ID, created.ProjectID)
	}

	if created.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, created.UserID)
	}

	if created.Action != string(models.ActionCreated) {
		t.Errorf("Expected action 'created', got '%s'", created.Action)
	}

	if created.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestActivityRepository_ListByProjectID(t *testing.T) {
	activityRepo, userRepo, projectRepo, _, cleanup := setupActivityRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForActivity(t, projectRepo, userRepo)

	// Create multiple activities
	for i := 0; i < 3; i++ {
		activity := &models.Activity{
			ProjectID:  &project.ID,
			UserID:     user.ID,
			Action:     string(models.ActionCreated),
			EntityType: string(models.EntityTypeIssue),
		}

		_, err := activityRepo.Create(ctx, activity)
		if err != nil {
			t.Fatalf("Failed to create activity %d: %v", i, err)
		}
	}

	// List activities
	activities, err := activityRepo.ListByProjectID(ctx, project.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(activities) != 3 {
		t.Errorf("Expected 3 activities, got %d", len(activities))
	}

	// Verify User field is populated (JOIN)
	for _, activity := range activities {
		if activity.User == nil {
			t.Error("Expected User to be populated, got nil")
		} else if activity.User.Email == "" {
			t.Error("Expected User.Email to be set")
		}
	}

	// Verify ordering (most recent first)
	if len(activities) >= 2 {
		if activities[0].CreatedAt.Before(activities[1].CreatedAt) {
			t.Error("Expected activities to be ordered by created_at DESC")
		}
	}
}

func TestActivityRepository_ListByIssueID(t *testing.T) {
	activityRepo, userRepo, projectRepo, issueRepo, cleanup := setupActivityRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForActivity(t, projectRepo, userRepo)

	// Create board column first
	boardRepo := NewBoardRepository(setupTestDB(t))
	err := boardRepo.CreateDefaultColumns(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to create default columns: %v", err)
	}

	columns, err := boardRepo.ListByProjectID(ctx, project.ID)
	if err != nil || len(columns) == 0 {
		t.Fatalf("Failed to get board columns: %v", err)
	}

	// Create an issue
	issue := &models.Issue{
		ProjectID:   project.ID,
		Title:       "Test Issue",
		Description: stringPtr("Test description"),
		ReporterID:  user.ID,
		ColumnID:    &columns[0].ID,
		Priority:    models.PriorityMedium,
		Status:      models.IssueStatusOpen,
	}

	createdIssue, err := issueRepo.Create(ctx, issue)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Create activities for this issue
	for i := 0; i < 2; i++ {
		activity := &models.Activity{
			ProjectID:  &project.ID,
			IssueID:    &createdIssue.ID,
			UserID:     user.ID,
			Action:     string(models.ActionUpdated),
			EntityType: string(models.EntityTypeIssue),
			EntityID:   &createdIssue.ID,
		}

		_, err := activityRepo.Create(ctx, activity)
		if err != nil {
			t.Fatalf("Failed to create activity %d: %v", i, err)
		}
	}

	// List activities by issue
	activities, err := activityRepo.ListByIssueID(ctx, createdIssue.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByIssueID failed: %v", err)
	}

	if len(activities) != 2 {
		t.Errorf("Expected 2 activities, got %d", len(activities))
	}

	// Verify all activities belong to the issue
	for _, activity := range activities {
		if activity.IssueID == nil || *activity.IssueID != createdIssue.ID {
			t.Errorf("Expected IssueID %d, got %v", createdIssue.ID, activity.IssueID)
		}
	}
}

func TestActivityRepository_ListWithFilter(t *testing.T) {
	activityRepo, userRepo, projectRepo, _, cleanup := setupActivityRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForActivity(t, projectRepo, userRepo)

	// Create activities with different actions
	actions := []models.ActivityAction{
		models.ActionCreated,
		models.ActionUpdated,
		models.ActionDeleted,
	}

	for _, action := range actions {
		activity := &models.Activity{
			ProjectID:  &project.ID,
			UserID:     user.ID,
			Action:     string(action),
			EntityType: string(models.EntityTypeIssue),
		}

		_, err := activityRepo.Create(ctx, activity)
		if err != nil {
			t.Fatalf("Failed to create activity: %v", err)
		}
	}

	// Filter by action
	actionFilter := string(models.ActionCreated)
	filter := &models.ActivityFilter{
		ProjectID: &project.ID,
		Action:    &actionFilter,
		Limit:     10,
		Offset:    0,
	}

	activities, err := activityRepo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(activities) != 1 {
		t.Errorf("Expected 1 activity, got %d", len(activities))
	}

	if activities[0].Action != string(models.ActionCreated) {
		t.Errorf("Expected action 'created', got '%s'", activities[0].Action)
	}
}

func TestActivityRepository_ListWithPagination(t *testing.T) {
	activityRepo, userRepo, projectRepo, _, cleanup := setupActivityRepo(t)
	defer cleanup()

	ctx := context.Background()
	project, user := createTestProjectForActivity(t, projectRepo, userRepo)

	// Create 10 activities
	for i := 0; i < 10; i++ {
		activity := &models.Activity{
			ProjectID:  &project.ID,
			UserID:     user.ID,
			Action:     string(models.ActionCreated),
			EntityType: string(models.EntityTypeIssue),
		}

		_, err := activityRepo.Create(ctx, activity)
		if err != nil {
			t.Fatalf("Failed to create activity %d: %v", i, err)
		}
	}

	// First page (5 items)
	activities, err := activityRepo.ListByProjectID(ctx, project.ID, 5, 0)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(activities) != 5 {
		t.Errorf("Expected 5 activities on first page, got %d", len(activities))
	}

	// Second page (5 items)
	activities2, err := activityRepo.ListByProjectID(ctx, project.ID, 5, 5)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(activities2) != 5 {
		t.Errorf("Expected 5 activities on second page, got %d", len(activities2))
	}

	// Verify no overlap
	if activities[0].ID == activities2[0].ID {
		t.Error("Expected different activities on different pages")
	}
}
