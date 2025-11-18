package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/internal/service"
	"github.com/yourusername/issue-tracker/pkg/database"
)

func setupProjectHandler(t *testing.T) (*ProjectHandler, *repository.UserRepository, *service.ProjectService, func()) {
	db, err := database.NewPostgresDB(database.Config{
		URL: "postgres://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	projectRepo := repository.NewProjectRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	userRepo := repository.NewUserRepository(db)
	projectService := service.NewProjectService(projectRepo, boardRepo, db)
	projectHandler := NewProjectHandler(projectService)

	cleanup := func() {
		db.Exec("DELETE FROM project_members WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'HDL%')")
		db.Exec("DELETE FROM board_columns WHERE project_id IN (SELECT id FROM projects WHERE key LIKE 'HDL%')")
		db.Exec("DELETE FROM projects WHERE key LIKE 'HDL%'")
		db.Exec("DELETE FROM users WHERE email LIKE 'handlertest%@example.com'")
		db.Close()
	}

	return projectHandler, userRepo, projectService, cleanup
}

func TestProjectHandler_Create(t *testing.T) {
	handler, userRepo, _, cleanup := setupProjectHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "handlertest1@example.com",
		Username:     "handlertest1",
		PasswordHash: "hash",
	}
	user, _ := userRepo.Create(ctx, testUser)

	t.Run("should create project successfully", func(t *testing.T) {
		desc := "Handler test project"
		reqBody := models.CreateProjectRequest{
			Name:        "Handler Test Project",
			Key:         "HDL1",
			Description: &desc,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Add user ID to context (simulating auth middleware)
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Project
		json.NewDecoder(w.Body).Decode(&response)

		if response.Name != reqBody.Name {
			t.Errorf("Expected name %s, got %s", reqBody.Name, response.Name)
		}

		if response.Key != reqBody.Key {
			t.Errorf("Expected key %s, got %s", reqBody.Key, response.Key)
		}
	})

	t.Run("should return bad request for invalid input", func(t *testing.T) {
		reqBody := models.CreateProjectRequest{
			Name: "", // Invalid: empty name
			Key:  "HDL2",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestProjectHandler_GetByID(t *testing.T) {
	handler, userRepo, projectService, cleanup := setupProjectHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	testUser := &models.User{
		Email:        "handlertest2@example.com",
		Username:     "handlertest2",
		PasswordHash: "hash",
	}
	user, _ := userRepo.Create(ctx, testUser)

	// Create project
	req := &models.CreateProjectRequest{
		Name: "Get Test Project",
		Key:  "HDL3",
	}
	project, _ := projectService.Create(ctx, req, user.ID)

	t.Run("should get project by ID", func(t *testing.T) {
		idStr := strconv.Itoa(project.ID)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+idStr, nil)
		req.SetPathValue("id", idStr)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.GetByID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Project
		json.NewDecoder(w.Body).Decode(&response)

		if response.ID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, response.ID)
		}
	})
}

func TestProjectHandler_List(t *testing.T) {
	handler, userRepo, projectService, cleanup := setupProjectHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "handlertest3@example.com",
		Username:     "handlertest3",
		PasswordHash: "hash",
	}
	user, _ := userRepo.Create(ctx, testUser)

	// Create projects
	req1 := &models.CreateProjectRequest{Name: "List Project 1", Key: "HDL4"}
	req2 := &models.CreateProjectRequest{Name: "List Project 2", Key: "HDL5"}
	projectService.Create(ctx, req1, user.ID)
	projectService.Create(ctx, req2, user.ID)

	t.Run("should list projects for user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []*models.Project
		json.NewDecoder(w.Body).Decode(&response)

		if len(response) < 2 {
			t.Errorf("Expected at least 2 projects, got %d", len(response))
		}
	})
}

func TestProjectHandler_Update(t *testing.T) {
	handler, userRepo, projectService, cleanup := setupProjectHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "handlertest4@example.com",
		Username:     "handlertest4",
		PasswordHash: "hash",
	}
	user, _ := userRepo.Create(ctx, testUser)

	// Create project
	req := &models.CreateProjectRequest{
		Name: "Update Test Project",
		Key:  "HDL6",
	}
	project, _ := projectService.Create(ctx, req, user.ID)

	t.Run("should update project", func(t *testing.T) {
		newName := "Updated Project Name"
		updateReq := models.UpdateProjectRequest{
			Name: &newName,
		}

		idStr := strconv.Itoa(project.ID)
		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+idStr, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", idStr)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Update(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Project
		json.NewDecoder(w.Body).Decode(&response)

		if response.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, response.Name)
		}
	})
}

func TestProjectHandler_Delete(t *testing.T) {
	handler, userRepo, projectService, cleanup := setupProjectHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create test user
	testUser := &models.User{
		Email:        "handlertest5@example.com",
		Username:     "handlertest5",
		PasswordHash: "hash",
	}
	user, _ := userRepo.Create(ctx, testUser)

	// Create project
	req := &models.CreateProjectRequest{
		Name: "Delete Test Project",
		Key:  "HDL7",
	}
	project, _ := projectService.Create(ctx, req, user.ID)

	t.Run("should delete project", func(t *testing.T) {
		idStr := strconv.Itoa(project.ID)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+idStr, nil)
		req.SetPathValue("id", idStr)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})
}
