package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupProjectRepo(t *testing.T) (*ProjectRepository, *UserRepository, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	projectRepo := NewProjectRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM projects WHERE key LIKE 'TEST%'")
		db.Exec("DELETE FROM users WHERE email LIKE 'projecttest%@example.com'")
		db.Close()
	}

	return projectRepo, userRepo, cleanup
}

func TestProjectRepository_Create(t *testing.T) {
	repo, userRepo, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "projecttest1@example.com",
		Username:     "projecttest1",
		PasswordHash: "hash",
	}
	createdUser, err := userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("should create project successfully", func(t *testing.T) {
		desc := "Test project description"
		project := &models.Project{
			Name:        "Test Project",
			Key:         "TEST1",
			Description: &desc,
			OwnerID:     createdUser.ID,
		}

		created, err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if created.ID == 0 {
			t.Error("Expected ID to be set")
		}

		if created.Name != project.Name {
			t.Errorf("Expected name %s, got %s", project.Name, created.Name)
		}

		if created.Key != project.Key {
			t.Errorf("Expected key %s, got %s", project.Key, created.Key)
		}

		if created.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("should return error for duplicate key", func(t *testing.T) {
		project1 := &models.Project{
			Name:    "Test Project 2",
			Key:     "TEST2",
			OwnerID: createdUser.ID,
		}

		_, err := repo.Create(ctx, project1)
		if err != nil {
			t.Fatalf("Failed to create first project: %v", err)
		}

		// Try to create another project with same key
		project2 := &models.Project{
			Name:    "Test Project 3",
			Key:     "TEST2",
			OwnerID: createdUser.ID,
		}

		_, err = repo.Create(ctx, project2)
		if err != pkgerrors.ErrConflict {
			t.Errorf("Expected ErrConflict, got %v", err)
		}
	})
}

func TestProjectRepository_GetByID(t *testing.T) {
	repo, userRepo, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "projecttest2@example.com",
		Username:     "projecttest2",
		PasswordHash: "hash",
	}
	createdUser, _ := userRepo.Create(ctx, testUser)

	t.Run("should get project by ID", func(t *testing.T) {
		project := &models.Project{
			Name:    "Test Project 4",
			Key:     "TEST4",
			OwnerID: createdUser.ID,
		}

		created, err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		found, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
		}

		if found.Name != project.Name {
			t.Errorf("Expected name %s, got %s", project.Name, found.Name)
		}
	})

	t.Run("should return ErrNotFound for non-existent ID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999999)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestProjectRepository_ListByUserID(t *testing.T) {
	repo, userRepo, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	testUser1 := &models.User{
		Email:        "projecttest3@example.com",
		Username:     "projecttest3",
		PasswordHash: "hash",
	}
	user1, _ := userRepo.Create(ctx, testUser1)

	testUser2 := &models.User{
		Email:        "projecttest4@example.com",
		Username:     "projecttest4",
		PasswordHash: "hash",
	}
	user2, _ := userRepo.Create(ctx, testUser2)

	t.Run("should list projects by user ID", func(t *testing.T) {
		// Create projects with different owners
		project1 := &models.Project{
			Name:    "User 1 Project 1",
			Key:     "TEST5",
			OwnerID: user1.ID,
		}

		project2 := &models.Project{
			Name:    "User 1 Project 2",
			Key:     "TEST6",
			OwnerID: user1.ID,
		}

		project3 := &models.Project{
			Name:    "User 2 Project",
			Key:     "TEST7",
			OwnerID: user2.ID,
		}

		repo.Create(ctx, project1)
		repo.Create(ctx, project2)
		repo.Create(ctx, project3)

		// Get projects for user 1
		projects, err := repo.ListByUserID(ctx, user1.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(projects) < 2 {
			t.Errorf("Expected at least 2 projects, got %d", len(projects))
		}

		// Verify all projects belong to user 1
		for _, p := range projects {
			if p.OwnerID != user1.ID {
				t.Errorf("Expected owner ID %d, got %d", user1.ID, p.OwnerID)
			}
		}
	})
}

func TestProjectRepository_Update(t *testing.T) {
	repo, userRepo, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "projecttest5@example.com",
		Username:     "projecttest5",
		PasswordHash: "hash",
	}
	createdUser, _ := userRepo.Create(ctx, testUser)

	t.Run("should update project", func(t *testing.T) {
		project := &models.Project{
			Name:    "Original Name",
			Key:     "TEST8",
			OwnerID: createdUser.ID,
		}

		created, err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		// Update project
		newDesc := "Updated description"
		created.Name = "Updated Name"
		created.Description = &newDesc

		err = repo.Update(ctx, created)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify update
		updated, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get updated project: %v", err)
		}

		if updated.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
		}

		if updated.Description == nil || *updated.Description != newDesc {
			t.Error("Expected description to be updated")
		}
	})
}

func TestProjectRepository_Delete(t *testing.T) {
	repo, userRepo, cleanup := setupProjectRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "projecttest6@example.com",
		Username:     "projecttest6",
		PasswordHash: "hash",
	}
	createdUser, _ := userRepo.Create(ctx, testUser)

	t.Run("should delete project", func(t *testing.T) {
		project := &models.Project{
			Name:    "To Delete",
			Key:     "TEST9",
			OwnerID: createdUser.ID,
		}

		created, err := repo.Create(ctx, project)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		// Delete project
		err = repo.Delete(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify deletion
		_, err = repo.GetByID(ctx, created.ID)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound after deletion, got %v", err)
		}
	})
}
