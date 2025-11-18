package service

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupBoardService(t *testing.T) (*BoardService, *repository.ProjectRepository, *repository.UserRepository, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Clean up test data
	db.Exec("DELETE FROM board_columns")
	db.Exec("DELETE FROM project_members")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM users")

	boardRepo := repository.NewBoardRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	userRepo := repository.NewUserRepository(db)

	boardService := NewBoardService(boardRepo, projectRepo, db)

	cleanup := func() {
		db.Exec("DELETE FROM board_columns")
		db.Exec("DELETE FROM project_members")
		db.Exec("DELETE FROM projects")
		db.Exec("DELETE FROM users")
	}

	return boardService, projectRepo, userRepo, cleanup
}

func createTestProjectForBoardService(t *testing.T, projectRepo *repository.ProjectRepository, userRepo *repository.UserRepository) (*models.Project, int) {
	ctx := context.Background()

	// Create test user
	user := &models.User{
		Email:        "boardservicetest@example.com",
		Username:     "boardservicetest",
		PasswordHash: "hash",
	}
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	project := &models.Project{
		Name:    "Board Service Test",
		Key:     "BSTEST",
		OwnerID: createdUser.ID,
	}
	createdProject, err := projectRepo.Create(ctx, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Add user as project member
	db, _ := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	_, err = db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, createdProject.ID, createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to add project member: %v", err)
	}

	return createdProject, createdUser.ID
}

func TestBoardService_List(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, userID := createTestProjectForBoardService(t, projectRepo, userRepo)

	// Create some columns first
	req1 := &models.CreateBoardColumnRequest{Name: "Backlog", Position: 0}
	req2 := &models.CreateBoardColumnRequest{Name: "In Progress", Position: 1}

	service.CreateColumn(ctx, project.ID, req1, userID)
	service.CreateColumn(ctx, project.ID, req2, userID)

	// List columns
	columns, err := service.List(ctx, project.ID, userID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(columns))
	}
}

func TestBoardService_List_NoPermission(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForBoardService(t, projectRepo, userRepo)

	// Create different user (not a member)
	otherUser := &models.User{
		Email:        "boardother@example.com",
		Username:     "boardother",
		PasswordHash: "hash",
	}
	createdOther, err := userRepo.Create(ctx, otherUser)
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	_, err = service.List(ctx, project.ID, createdOther.ID)
	if err != pkgerrors.ErrForbidden {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestBoardService_CreateColumn(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, userID := createTestProjectForBoardService(t, projectRepo, userRepo)

	req := &models.CreateBoardColumnRequest{
		Name:     "Testing",
		Position: 0,
	}

	column, err := service.CreateColumn(ctx, project.ID, req, userID)
	if err != nil {
		t.Fatalf("CreateColumn failed: %v", err)
	}

	if column.Name != "Testing" {
		t.Errorf("Expected name 'Testing', got '%s'", column.Name)
	}
}

func TestBoardService_CreateColumn_NoPermission(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, _ := createTestProjectForBoardService(t, projectRepo, userRepo)

	// Create different user (not owner/admin)
	otherUser := &models.User{
		Email:        "boardmember@example.com",
		Username:     "boardmember",
		PasswordHash: "hash",
	}
	createdOther, err := userRepo.Create(ctx, otherUser)
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	// Add as regular member
	db, _ := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, project.ID, createdOther.ID)

	req := &models.CreateBoardColumnRequest{
		Name:     "Testing",
		Position: 0,
	}

	_, err = service.CreateColumn(ctx, project.ID, req, createdOther.ID)
	if err != pkgerrors.ErrForbidden {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestBoardService_UpdateColumn(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, userID := createTestProjectForBoardService(t, projectRepo, userRepo)

	createReq := &models.CreateBoardColumnRequest{
		Name:     "Old Name",
		Position: 0,
	}
	created, _ := service.CreateColumn(ctx, project.ID, createReq, userID)

	newName := "New Name"
	newPosition := 1
	updateReq := &models.UpdateBoardColumnRequest{
		Name:     &newName,
		Position: &newPosition,
	}

	updated, err := service.UpdateColumn(ctx, created.ID, updateReq, userID)
	if err != nil {
		t.Fatalf("UpdateColumn failed: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}
	if updated.Position != 1 {
		t.Errorf("Expected position 1, got %d", updated.Position)
	}
}

func TestBoardService_DeleteColumn(t *testing.T) {
	service, projectRepo, userRepo, cleanup := setupBoardService(t)
	defer cleanup()

	ctx := context.Background()
	project, userID := createTestProjectForBoardService(t, projectRepo, userRepo)

	createReq := &models.CreateBoardColumnRequest{
		Name:     "To Delete",
		Position: 0,
	}
	created, _ := service.CreateColumn(ctx, project.ID, createReq, userID)

	err := service.DeleteColumn(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("DeleteColumn failed: %v", err)
	}

	// Verify deleted
	columns, _ := service.List(ctx, project.ID, userID)
	if len(columns) != 0 {
		t.Error("Expected column to be deleted")
	}
}
