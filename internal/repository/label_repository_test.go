package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupLabelRepo(t *testing.T) (*LabelRepository, *ProjectRepository, *UserRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM issue_labels")
	db.Exec("DELETE FROM labels")
	db.Exec("DELETE FROM issues")
	db.Exec("DELETE FROM project_members")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	labelRepo := NewLabelRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM issue_labels")
		db.Exec("DELETE FROM labels")
		db.Exec("DELETE FROM issues")
		db.Exec("DELETE FROM project_members")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return labelRepo, projectRepo, userRepo, cleanup
}

func createTestProject(t *testing.T, projectRepo *ProjectRepository, userRepo *UserRepository) *models.Project {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Email:        "labeltest@example.com",
		PasswordHash: "hash",
	}
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	project := &models.Project{
		Name:    "Test Project",
		Key:     "TEST",
		OwnerID: createdUser.ID,
	}
	createdProject, err := projectRepo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return createdProject
}

func TestLabelRepository_Create(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	label := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}

	created, err := labelRepo.Create(ctx, label)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if created.Name != "bug" {
		t.Errorf("Expected name 'bug', got '%s'", created.Name)
	}
	if created.Color != "#ff0000" {
		t.Errorf("Expected color '#ff0000', got '%s'", created.Color)
	}
}

func TestLabelRepository_Create_DuplicateName(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	label1 := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}
	_, err := labelRepo.Create(ctx, label1)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Try to create duplicate
	label2 := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#00ff00",
	}
	_, err = labelRepo.Create(ctx, label2)
	if err == nil {
		t.Error("Expected error for duplicate label name")
	}
}

func TestLabelRepository_GetByID(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	label := &models.Label{
		ProjectID: project.ID,
		Name:      "feature",
		Color:     "#00ff00",
	}
	created, _ := labelRepo.Create(ctx, label)

	found, err := labelRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Name != "feature" {
		t.Errorf("Expected name 'feature', got '%s'", found.Name)
	}
}

func TestLabelRepository_GetByID_NotFound(t *testing.T) {
	labelRepo, _, _, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()

	_, err := labelRepo.GetByID(ctx, 99999)
	if err != pkgerrors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestLabelRepository_ListByProjectID(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	// Create multiple labels
	labels := []*models.Label{
		{ProjectID: project.ID, Name: "bug", Color: "#ff0000"},
		{ProjectID: project.ID, Name: "feature", Color: "#00ff00"},
		{ProjectID: project.ID, Name: "enhancement", Color: "#0000ff"},
	}

	for _, l := range labels {
		_, err := labelRepo.Create(ctx, l)
		if err != nil {
			t.Fatalf("Failed to create label: %v", err)
		}
	}

	found, err := labelRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(found))
	}
}

func TestLabelRepository_Update(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	label := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}
	created, _ := labelRepo.Create(ctx, label)

	// Update
	created.Name = "critical-bug"
	created.Color = "#ff00ff"

	err := labelRepo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	updated, _ := labelRepo.GetByID(ctx, created.ID)
	if updated.Name != "critical-bug" {
		t.Errorf("Expected name 'critical-bug', got '%s'", updated.Name)
	}
	if updated.Color != "#ff00ff" {
		t.Errorf("Expected color '#ff00ff', got '%s'", updated.Color)
	}
}

func TestLabelRepository_Delete(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	label := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}
	created, _ := labelRepo.Create(ctx, label)

	err := labelRepo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = labelRepo.GetByID(ctx, created.ID)
	if err != pkgerrors.ErrNotFound {
		t.Error("Expected label to be deleted")
	}
}

func TestLabelRepository_AddToIssue(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	// Create label
	label := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}
	createdLabel, _ := labelRepo.Create(ctx, label)

	// Create issue
	db := setupTestDB(t)
	var issueID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO issues (project_id, issue_number, title, status, priority, reporter_id)
		VALUES ($1, 1, 'Test issue', 'open', 'medium', $2)
		RETURNING id
	`, project.ID, project.OwnerID).Scan(&issueID)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Add label to issue
	err = labelRepo.AddToIssue(ctx, issueID, createdLabel.ID)
	if err != nil {
		t.Fatalf("AddToIssue failed: %v", err)
	}

	// Verify
	labels, err := labelRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		t.Fatalf("ListByIssueID failed: %v", err)
	}

	if len(labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(labels))
	}
	if len(labels) > 0 && labels[0].Name != "bug" {
		t.Errorf("Expected label name 'bug', got '%s'", labels[0].Name)
	}
}

func TestLabelRepository_RemoveFromIssue(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	// Create label
	label := &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	}
	createdLabel, _ := labelRepo.Create(ctx, label)

	// Create issue
	db := setupTestDB(t)
	var issueID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO issues (project_id, issue_number, title, status, priority, reporter_id)
		VALUES ($1, 1, 'Test issue', 'open', 'medium', $2)
		RETURNING id
	`, project.ID, project.OwnerID).Scan(&issueID)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Add label to issue
	labelRepo.AddToIssue(ctx, issueID, createdLabel.ID)

	// Remove label from issue
	err = labelRepo.RemoveFromIssue(ctx, issueID, createdLabel.ID)
	if err != nil {
		t.Fatalf("RemoveFromIssue failed: %v", err)
	}

	// Verify
	labels, _ := labelRepo.ListByIssueID(ctx, issueID)
	if len(labels) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(labels))
	}
}

func TestLabelRepository_ListByIssueID(t *testing.T) {
	labelRepo, projectRepo, userRepo, cleanup := setupLabelRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProject(t, projectRepo, userRepo)

	// Create multiple labels
	label1, _ := labelRepo.Create(ctx, &models.Label{
		ProjectID: project.ID,
		Name:      "bug",
		Color:     "#ff0000",
	})
	label2, _ := labelRepo.Create(ctx, &models.Label{
		ProjectID: project.ID,
		Name:      "feature",
		Color:     "#00ff00",
	})

	// Create issue
	db := setupTestDB(t)
	var issueID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO issues (project_id, issue_number, title, status, priority, reporter_id)
		VALUES ($1, 1, 'Test issue', 'open', 'medium', $2)
		RETURNING id
	`, project.ID, project.OwnerID).Scan(&issueID)
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	// Add multiple labels
	labelRepo.AddToIssue(ctx, issueID, label1.ID)
	labelRepo.AddToIssue(ctx, issueID, label2.ID)

	// List labels
	labels, err := labelRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		t.Fatalf("ListByIssueID failed: %v", err)
	}

	if len(labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(labels))
	}
}
