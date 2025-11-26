package repository

import (
	"context"
	"testing"

	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

func TestWebhookRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-test@example.com",
		Username: "webhooktest",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "Webhook Test Project",
		Key:     "WTP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a webhook
	webhook := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Test Webhook",
		URL:       "https://example.com/webhook",
		Events:    []string{models.EventIssueCreated, models.EventIssueUpdated},
		IsActive:  true,
		CreatedBy: user.ID,
	}

	created, err := repo.Create(context.Background(), webhook)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected webhook ID to be set")
	}
	if created.Name != "Test Webhook" {
		t.Errorf("Expected name 'Test Webhook', got '%s'", created.Name)
	}
	if created.URL != "https://example.com/webhook" {
		t.Errorf("Expected URL 'https://example.com/webhook', got '%s'", created.URL)
	}
	if len(created.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(created.Events))
	}
	if !created.IsActive {
		t.Error("Expected webhook to be active")
	}
}

func TestWebhookRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-get@example.com",
		Username: "webhookget",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "Get Test Project",
		Key:     "GTP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a webhook
	webhook := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Get Test Webhook",
		URL:       "https://example.com/get",
		Events:    []string{models.EventIssueCreated},
		IsActive:  true,
		CreatedBy: user.ID,
	}

	created, err := repo.Create(context.Background(), webhook)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Get the webhook
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get webhook: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
	if retrieved.Name != "Get Test Webhook" {
		t.Errorf("Expected name 'Get Test Webhook', got '%s'", retrieved.Name)
	}
}

func TestWebhookRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)

	_, err := repo.GetByID(context.Background(), 99999)
	if err != pkgerrors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestWebhookRepository_ListByProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-list@example.com",
		Username: "webhooklist",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "List Test Project",
		Key:     "LTP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create multiple webhooks
	for i := 1; i <= 3; i++ {
		webhook := &models.Webhook{
			ProjectID: project.ID,
			Name:      "Webhook " + string(rune('0'+i)),
			URL:       "https://example.com/hook" + string(rune('0'+i)),
			Events:    []string{models.EventIssueCreated},
			IsActive:  true,
			CreatedBy: user.ID,
		}
		_, err := repo.Create(context.Background(), webhook)
		if err != nil {
			t.Fatalf("Failed to create webhook %d: %v", i, err)
		}
	}

	// List webhooks
	webhooks, err := repo.ListByProject(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}

	if len(webhooks) != 3 {
		t.Errorf("Expected 3 webhooks, got %d", len(webhooks))
	}
}

func TestWebhookRepository_ListActiveByProjectAndEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-event@example.com",
		Username: "webhookevent",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "Event Test Project",
		Key:     "ETP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create webhooks with different events
	webhook1 := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Issue Webhook",
		URL:       "https://example.com/issue",
		Events:    []string{models.EventIssueCreated, models.EventIssueUpdated},
		IsActive:  true,
		CreatedBy: user.ID,
	}
	_, err = repo.Create(context.Background(), webhook1)
	if err != nil {
		t.Fatalf("Failed to create webhook1: %v", err)
	}

	webhook2 := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Comment Webhook",
		URL:       "https://example.com/comment",
		Events:    []string{models.EventCommentCreated},
		IsActive:  true,
		CreatedBy: user.ID,
	}
	_, err = repo.Create(context.Background(), webhook2)
	if err != nil {
		t.Fatalf("Failed to create webhook2: %v", err)
	}

	// Create an inactive webhook
	webhook3 := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Inactive Webhook",
		URL:       "https://example.com/inactive",
		Events:    []string{models.EventIssueCreated},
		IsActive:  false,
		CreatedBy: user.ID,
	}
	_, err = repo.Create(context.Background(), webhook3)
	if err != nil {
		t.Fatalf("Failed to create webhook3: %v", err)
	}

	// Query for issue.created event
	webhooks, err := repo.ListActiveByProjectAndEvent(context.Background(), project.ID, models.EventIssueCreated)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}

	if len(webhooks) != 1 {
		t.Errorf("Expected 1 active webhook for issue.created, got %d", len(webhooks))
	}

	// Query for comment.created event
	webhooks, err = repo.ListActiveByProjectAndEvent(context.Background(), project.ID, models.EventCommentCreated)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}

	if len(webhooks) != 1 {
		t.Errorf("Expected 1 active webhook for comment.created, got %d", len(webhooks))
	}
}

func TestWebhookRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-update@example.com",
		Username: "webhookupdate",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "Update Test Project",
		Key:     "UTP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a webhook
	webhook := &models.Webhook{
		ProjectID: project.ID,
		Name:      "Original Name",
		URL:       "https://example.com/original",
		Events:    []string{models.EventIssueCreated},
		IsActive:  true,
		CreatedBy: user.ID,
	}

	created, err := repo.Create(context.Background(), webhook)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Update the webhook
	created.Name = "Updated Name"
	created.URL = "https://example.com/updated"
	created.IsActive = false

	err = repo.Update(context.Background(), created)
	if err != nil {
		t.Fatalf("Failed to update webhook: %v", err)
	}

	// Verify update
	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated webhook: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.URL != "https://example.com/updated" {
		t.Errorf("Expected URL 'https://example.com/updated', got '%s'", updated.URL)
	}
	if updated.IsActive {
		t.Error("Expected webhook to be inactive")
	}
}

func TestWebhookRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewWebhookRepository(db)
	userRepo := NewUserRepository(db)
	projectRepo := NewProjectRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "webhook-delete@example.com",
		Username: "webhookdelete",
		PasswordHash: "hashedpassword123",
	}
	user, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test project
	project := &models.Project{
		Name:    "Delete Test Project",
		Key:     "DTP",
		OwnerID: user.ID,
	}
	project, err = projectRepo.Create(context.Background(), project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a webhook
	webhook := &models.Webhook{
		ProjectID: project.ID,
		Name:      "To Delete",
		URL:       "https://example.com/delete",
		Events:    []string{models.EventIssueCreated},
		IsActive:  true,
		CreatedBy: user.ID,
	}

	created, err := repo.Create(context.Background(), webhook)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Delete the webhook
	err = repo.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete webhook: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(context.Background(), created.ID)
	if err != pkgerrors.ErrNotFound {
		t.Errorf("Expected ErrNotFound after deletion, got %v", err)
	}
}
