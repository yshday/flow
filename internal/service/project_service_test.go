package service

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/database"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func setupProjectService(t *testing.T) (*ProjectService, *repository.UserRepository, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	projectRepo := repository.NewProjectRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	userRepo := repository.NewUserRepository(db)
	projectService := NewProjectService(projectRepo, boardRepo, db)

	cleanup := func() {
		db.Exec("DELETE FROM project_members WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'SVC%')")
		db.Exec("DELETE FROM board_columns WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'SVC%')")
		db.Exec("DELETE FROM projects WHERE key LIKE 'SVC%'")
		db.Exec("DELETE FROM users WHERE email LIKE 'svctest%@example.com'")
		db.Close()
	}

	return projectService, userRepo, cleanup
}

func TestProjectService_Create(t *testing.T) {
	service, userRepo, cleanup := setupProjectService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "svctest1@example.com",
		Username:     "svctest1",
		PasswordHash: "hash",
	}
	owner, err := userRepo.Create(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("should create project with default columns and owner membership", func(t *testing.T) {
		desc := "Service test project"
		req := &models.CreateProjectRequest{
			Name:        "Service Test Project",
			Key:         "SVC1",
			Description: &desc,
		}

		project, err := service.Create(ctx, req, owner.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if project.ID == 0 {
			t.Error("Expected project ID to be set")
		}

		if project.Name != req.Name {
			t.Errorf("Expected name %s, got %s", req.Name, project.Name)
		}

		if project.OwnerID != owner.ID {
			t.Errorf("Expected owner ID %d, got %d", owner.ID, project.OwnerID)
		}

		// Verify default columns were created
		boardRepo := repository.NewBoardRepository(service.db)
		columns, err := boardRepo.ListByProjectID(ctx, project.ID)
		if err != nil {
			t.Fatalf("Failed to get board columns: %v", err)
		}

		if len(columns) != 3 {
			t.Errorf("Expected 3 default columns, got %d", len(columns))
		}

		expectedColumns := []string{"Backlog", "In Progress", "Done"}
		for i, col := range columns {
			if col.Name != expectedColumns[i] {
				t.Errorf("Expected column %d to be %s, got %s", i, expectedColumns[i], col.Name)
			}
		}

		// Verify owner was added as member
		var role string
		err = service.db.QueryRowContext(ctx, `
			SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2
		`, project.ID, owner.ID).Scan(&role)
		if err != nil {
			t.Fatalf("Failed to query project member: %v", err)
		}

		if role != string(models.RoleOwner) {
			t.Errorf("Expected owner role, got %s", role)
		}
	})
}

func TestProjectService_GetByID(t *testing.T) {
	service, userRepo, cleanup := setupProjectService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	owner := &models.User{
		Email:        "svctest2@example.com",
		Username:     "svctest2",
		PasswordHash: "hash",
	}
	ownerCreated, _ := userRepo.Create(ctx, owner)

	otherUser := &models.User{
		Email:        "svctest3@example.com",
		Username:     "svctest3",
		PasswordHash: "hash",
	}
	otherUserCreated, _ := userRepo.Create(ctx, otherUser)

	// Create project
	req := &models.CreateProjectRequest{
		Name: "Get By ID Test",
		Key:  "SVC2",
	}
	project, _ := service.Create(ctx, req, ownerCreated.ID)

	t.Run("should get project when user has access", func(t *testing.T) {
		found, err := service.GetByID(ctx, project.ID, ownerCreated.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, found.ID)
		}
	})

	t.Run("should return forbidden when user has no access", func(t *testing.T) {
		_, err := service.GetByID(ctx, project.ID, otherUserCreated.ID)
		if err != pkgerrors.ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})

	t.Run("should return not found for non-existent project", func(t *testing.T) {
		_, err := service.GetByID(ctx, 999999, ownerCreated.ID)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestProjectService_List(t *testing.T) {
	service, userRepo, cleanup := setupProjectService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	user1 := &models.User{
		Email:        "svctest4@example.com",
		Username:     "svctest4",
		PasswordHash: "hash",
	}
	user1Created, _ := userRepo.Create(ctx, user1)

	user2 := &models.User{
		Email:        "svctest5@example.com",
		Username:     "svctest5",
		PasswordHash: "hash",
	}
	user2Created, _ := userRepo.Create(ctx, user2)

	t.Run("should list projects owned by user", func(t *testing.T) {
		// Create projects
		req1 := &models.CreateProjectRequest{Name: "User1 Project 1", Key: "SVC3"}
		req2 := &models.CreateProjectRequest{Name: "User1 Project 2", Key: "SVC4"}
		req3 := &models.CreateProjectRequest{Name: "User2 Project", Key: "SVC5"}

		service.Create(ctx, req1, user1Created.ID)
		service.Create(ctx, req2, user1Created.ID)
		service.Create(ctx, req3, user2Created.ID)

		// Get projects for user1
		projects, err := service.List(ctx, user1Created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(projects) < 2 {
			t.Errorf("Expected at least 2 projects, got %d", len(projects))
		}
	})

	t.Run("should list projects where user is member", func(t *testing.T) {
		// Create project owned by user2
		req := &models.CreateProjectRequest{Name: "User2 Project 2", Key: "SVC6"}
		project, _ := service.Create(ctx, req, user2Created.ID)

		// Add user1 as member
		_, err := service.db.ExecContext(ctx, `
			INSERT INTO project_members (project_id, user_id, role)
			VALUES ($1, $2, $3)
		`, project.ID, user1Created.ID, models.RoleMember)
		if err != nil {
			t.Fatalf("Failed to add member: %v", err)
		}

		// Get projects for user1
		projects, err := service.List(ctx, user1Created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should include the project where user1 is a member
		found := false
		for _, p := range projects {
			if p.ID == project.ID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find project where user is member")
		}
	})
}

func TestProjectService_Update(t *testing.T) {
	service, userRepo, cleanup := setupProjectService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	owner := &models.User{
		Email:        "svctest6@example.com",
		Username:     "svctest6",
		PasswordHash: "hash",
	}
	ownerCreated, _ := userRepo.Create(ctx, owner)

	admin := &models.User{
		Email:        "svctest7@example.com",
		Username:     "svctest7",
		PasswordHash: "hash",
	}
	adminCreated, _ := userRepo.Create(ctx, admin)

	member := &models.User{
		Email:        "svctest8@example.com",
		Username:     "svctest8",
		PasswordHash: "hash",
	}
	memberCreated, _ := userRepo.Create(ctx, member)

	// Create project
	req := &models.CreateProjectRequest{Name: "Update Test", Key: "SVC7"}
	project, _ := service.Create(ctx, req, ownerCreated.ID)

	// Add admin and member
	service.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)
	`, project.ID, adminCreated.ID, models.RoleAdmin)
	service.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)
	`, project.ID, memberCreated.ID, models.RoleMember)

	t.Run("should allow owner to update project", func(t *testing.T) {
		newName := "Updated Name"
		updateReq := &models.UpdateProjectRequest{
			Name: &newName,
		}

		updated, err := service.Update(ctx, project.ID, updateReq, ownerCreated.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updated.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, updated.Name)
		}
	})

	t.Run("should allow admin to update project", func(t *testing.T) {
		newDesc := "Updated by admin"
		updateReq := &models.UpdateProjectRequest{
			Description: &newDesc,
		}

		updated, err := service.Update(ctx, project.ID, updateReq, adminCreated.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updated.Description == nil || *updated.Description != newDesc {
			t.Error("Expected description to be updated")
		}
	})

	t.Run("should not allow member to update project", func(t *testing.T) {
		newName := "Hacked"
		updateReq := &models.UpdateProjectRequest{
			Name: &newName,
		}

		_, err := service.Update(ctx, project.ID, updateReq, memberCreated.ID)
		if err != pkgerrors.ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})
}

func TestProjectService_Delete(t *testing.T) {
	service, userRepo, cleanup := setupProjectService(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	owner := &models.User{
		Email:        "svctest9@example.com",
		Username:     "svctest9",
		PasswordHash: "hash",
	}
	ownerCreated, _ := userRepo.Create(ctx, owner)

	admin := &models.User{
		Email:        "svctest10@example.com",
		Username:     "svctest10",
		PasswordHash: "hash",
	}
	adminCreated, _ := userRepo.Create(ctx, admin)

	t.Run("should allow owner to delete project", func(t *testing.T) {
		req := &models.CreateProjectRequest{Name: "To Delete", Key: "SVC8"}
		project, _ := service.Create(ctx, req, ownerCreated.ID)

		err := service.Delete(ctx, project.ID, ownerCreated.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify deletion
		_, err = service.GetByID(ctx, project.ID, ownerCreated.ID)
		if err != pkgerrors.ErrNotFound {
			t.Errorf("Expected ErrNotFound after deletion, got %v", err)
		}
	})

	t.Run("should not allow admin to delete project", func(t *testing.T) {
		req := &models.CreateProjectRequest{Name: "Admin Cannot Delete", Key: "SVC9"}
		project, _ := service.Create(ctx, req, ownerCreated.ID)

		// Add admin
		service.db.ExecContext(ctx, `
			INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)
		`, project.ID, adminCreated.ID, models.RoleAdmin)

		err := service.Delete(ctx, project.ID, adminCreated.ID)
		if err != pkgerrors.ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})
}
