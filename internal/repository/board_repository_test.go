package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupBoardRepo(t *testing.T) (*BoardRepository, *ProjectRepository, *UserRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM board_columns")
	db.Exec("DELETE FROM project_members")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	boardRepo := NewBoardRepository(db)
	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM board_columns")
		db.Exec("DELETE FROM project_members")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return boardRepo, projectRepo, userRepo, cleanup
}

func createTestProjectForBoard(t *testing.T, projectRepo *ProjectRepository, userRepo *UserRepository) *models.Project {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Email:        "boardtest@example.com",
		Username:     "boardtest",
		PasswordHash: "hash",
	}
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	project := &models.Project{
		Name:    "Board Test Project",
		Key:     "BOARD",
		OwnerID: createdUser.ID,
	}
	createdProject, err := projectRepo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	return createdProject
}

func TestBoardRepository_CreateDefaultColumns(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	err := boardRepo.CreateDefaultColumns(ctx, project.ID)
	if err != nil {
		t.Fatalf("CreateDefaultColumns failed: %v", err)
	}

	// Verify default columns were created
	columns, err := boardRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 default columns, got %d", len(columns))
	}

	expectedNames := []string{"Backlog", "In Progress", "Done"}
	for i, col := range columns {
		if col.Name != expectedNames[i] {
			t.Errorf("Expected column %d to be '%s', got '%s'", i, expectedNames[i], col.Name)
		}
		if col.Position != i {
			t.Errorf("Expected column %d position to be %d, got %d", i, i, col.Position)
		}
	}
}

func TestBoardRepository_ListByProjectID(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	// Create some columns
	columns := []*models.BoardColumn{
		{ProjectID: project.ID, Name: "To Do", Position: 0},
		{ProjectID: project.ID, Name: "Doing", Position: 1},
		{ProjectID: project.ID, Name: "Done", Position: 2},
	}

	for _, col := range columns {
		_, err := boardRepo.Create(ctx, col)
		if err != nil {
			t.Fatalf("Failed to create column: %v", err)
		}
	}

	// List columns
	found, err := boardRepo.ListByProjectID(ctx, project.ID)
	if err != nil {
		t.Fatalf("ListByProjectID failed: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(found))
	}

	// Verify they're sorted by position
	for i, col := range found {
		if col.Position != i {
			t.Errorf("Expected position %d, got %d", i, col.Position)
		}
	}
}

func TestBoardRepository_Create(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	column := &models.BoardColumn{
		ProjectID: project.ID,
		Name:      "Testing",
		Position:  0,
	}

	created, err := boardRepo.Create(ctx, column)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if created.Name != "Testing" {
		t.Errorf("Expected name 'Testing', got '%s'", created.Name)
	}
	if created.Position != 0 {
		t.Errorf("Expected position 0, got %d", created.Position)
	}
}

func TestBoardRepository_Delete(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	column := &models.BoardColumn{
		ProjectID: project.ID,
		Name:      "To Delete",
		Position:  0,
	}
	created, _ := boardRepo.Create(ctx, column)

	err := boardRepo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's deleted
	columns, _ := boardRepo.ListByProjectID(ctx, project.ID)
	if len(columns) != 0 {
		t.Error("Expected column to be deleted")
	}
}

func TestBoardRepository_Delete_NotFound(t *testing.T) {
	boardRepo, _, _, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()

	err := boardRepo.Delete(ctx, 99999)
	if err != pkgerrors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBoardRepository_GetByID(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	column := &models.BoardColumn{
		ProjectID: project.ID,
		Name:      "Test Column",
		Position:  0,
	}
	created, _ := boardRepo.Create(ctx, column)

	found, err := boardRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Name != "Test Column" {
		t.Errorf("Expected name 'Test Column', got '%s'", found.Name)
	}
}

func TestBoardRepository_GetByID_NotFound(t *testing.T) {
	boardRepo, _, _, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()

	_, err := boardRepo.GetByID(ctx, 99999)
	if err != pkgerrors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBoardRepository_Update(t *testing.T) {
	boardRepo, projectRepo, userRepo, cleanup := setupBoardRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := createTestProjectForBoard(t, projectRepo, userRepo)

	column := &models.BoardColumn{
		ProjectID: project.ID,
		Name:      "Old Name",
		Position:  0,
	}
	created, _ := boardRepo.Create(ctx, column)

	// Update
	created.Name = "New Name"
	created.Position = 1

	err := boardRepo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	updated, _ := boardRepo.GetByID(ctx, created.ID)
	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}
	if updated.Position != 1 {
		t.Errorf("Expected position 1, got %d", updated.Position)
	}
}
