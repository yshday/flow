package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
)

var notificationTestCounter = 0

func setupNotificationRepo(t *testing.T) (*NotificationRepository, *UserRepository, func()) {
	db := setupTestDB(t)

	// Clean up test data
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM users")

	notificationRepo := NewNotificationRepository(db)
	userRepo := NewUserRepository(db)

	cleanup := func() {
		db.Exec("DELETE FROM notifications")
		db.Exec("DELETE FROM users")
	}

	return notificationRepo, userRepo, cleanup
}

func createTestUserForNotification(t *testing.T, userRepo *UserRepository, suffix string) *models.User {
	user := &models.User{
		Email:        "notif_user_" + suffix + "@example.com",
		Username:     "notifuser_" + suffix,
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return createdUser
}

func TestNotificationRepository_Create(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)
	actor := createTestUserForNotification(t, userRepo, suffix+"_actor")

	notification := &models.Notification{
		UserID:     user.ID,
		ActorID:    &actor.ID,
		EntityType: models.NotificationEntityIssue,
		EntityID:   123,
		Action:     models.NotificationActionCreated,
		Title:      "New issue created",
		Message:    stringPtr("Issue #123 was created"),
		Read:       false,
	}

	created, err := notificationRepo.Create(ctx, notification)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}

	if created.Title != "New issue created" {
		t.Errorf("Expected title 'New issue created', got '%s'", created.Title)
	}

	if created.Read != false {
		t.Errorf("Expected Read to be false, got %v", created.Read)
	}

	if created.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestNotificationRepository_GetByID(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	notification := &models.Notification{
		UserID:     user.ID,
		EntityType: models.NotificationEntityComment,
		EntityID:   456,
		Action:     models.NotificationActionCommented,
		Title:      "New comment",
		Read:       false,
	}

	created, err := notificationRepo.Create(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Get by ID
	retrieved, err := notificationRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.Title != "New comment" {
		t.Errorf("Expected title 'New comment', got '%s'", retrieved.Title)
	}
}

func TestNotificationRepository_ListByUserID(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	// Create multiple notifications
	for i := 1; i <= 3; i++ {
		notification := &models.Notification{
			UserID:     user.ID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   100 + i,
			Action:     models.NotificationActionCreated,
			Title:      fmt.Sprintf("Notification %d", i),
			Read:       false,
		}

		_, err := notificationRepo.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Failed to create notification %d: %v", i, err)
		}

		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// List notifications for user
	notifications, err := notificationRepo.ListByUserID(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}

	if len(notifications) != 3 {
		t.Errorf("Expected 3 notifications, got %d", len(notifications))
	}

	// Check ordering (most recent first)
	if notifications[0].Title != "Notification 3" {
		t.Errorf("Expected first notification to be 'Notification 3', got '%s'", notifications[0].Title)
	}
}

func TestNotificationRepository_ListUnreadByUserID(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	// Create 2 unread and 1 read notification
	for i := 1; i <= 3; i++ {
		notification := &models.Notification{
			UserID:     user.ID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   200 + i,
			Action:     models.NotificationActionCreated,
			Title:      fmt.Sprintf("Notification %d", i),
			Read:       i == 2, // Second one is read
		}

		created, err := notificationRepo.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Failed to create notification %d: %v", i, err)
		}

		if i == 2 {
			// Mark as read
			now := time.Now()
			created.ReadAt = &now
			_, err = notificationRepo.Update(ctx, created)
			if err != nil {
				t.Fatalf("Failed to mark notification as read: %v", err)
			}
		}
	}

	// List only unread notifications
	unreadNotifications, err := notificationRepo.ListUnreadByUserID(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListUnreadByUserID failed: %v", err)
	}

	if len(unreadNotifications) != 2 {
		t.Errorf("Expected 2 unread notifications, got %d", len(unreadNotifications))
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	notification := &models.Notification{
		UserID:     user.ID,
		EntityType: models.NotificationEntityIssue,
		EntityID:   789,
		Action:     models.NotificationActionAssigned,
		Title:      "Assigned to issue",
		Read:       false,
	}

	created, err := notificationRepo.Create(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Mark as read
	err = notificationRepo.MarkAsRead(ctx, []int{created.ID})
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	// Verify it's marked as read
	retrieved, err := notificationRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve notification: %v", err)
	}

	if !retrieved.Read {
		t.Error("Expected notification to be marked as read")
	}

	if retrieved.ReadAt == nil {
		t.Error("Expected ReadAt to be set")
	}
}

func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	// Create 3 unread notifications
	for i := 1; i <= 3; i++ {
		notification := &models.Notification{
			UserID:     user.ID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   300 + i,
			Action:     models.NotificationActionCreated,
			Title:      fmt.Sprintf("Notification %d", i),
			Read:       false,
		}

		_, err := notificationRepo.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Failed to create notification %d: %v", i, err)
		}
	}

	// Mark all as read
	err := notificationRepo.MarkAllAsRead(ctx, user.ID)
	if err != nil {
		t.Fatalf("MarkAllAsRead failed: %v", err)
	}

	// Verify all are marked as read
	unreadNotifications, err := notificationRepo.ListUnreadByUserID(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list unread notifications: %v", err)
	}

	if len(unreadNotifications) != 0 {
		t.Errorf("Expected 0 unread notifications, got %d", len(unreadNotifications))
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	notification := &models.Notification{
		UserID:     user.ID,
		EntityType: models.NotificationEntityProject,
		EntityID:   999,
		Action:     models.NotificationActionDeleted,
		Title:      "Project deleted",
		Read:       false,
	}

	created, err := notificationRepo.Create(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Delete
	err = notificationRepo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = notificationRepo.GetByID(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after deleting notification, got nil")
	}
}

func TestNotificationRepository_CountUnread(t *testing.T) {
	notificationRepo, userRepo, cleanup := setupNotificationRepo(t)
	defer cleanup()

	ctx := context.Background()
	notificationTestCounter++
	suffix := fmt.Sprintf("%d", notificationTestCounter)

	user := createTestUserForNotification(t, userRepo, suffix)

	// Create 5 unread and 2 read notifications
	for i := 1; i <= 7; i++ {
		notification := &models.Notification{
			UserID:     user.ID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   400 + i,
			Action:     models.NotificationActionCreated,
			Title:      fmt.Sprintf("Notification %d", i),
			Read:       i <= 2, // First 2 are read
		}

		_, err := notificationRepo.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Failed to create notification %d: %v", i, err)
		}
	}

	// Count unread
	count, err := notificationRepo.CountUnread(ctx, user.ID)
	if err != nil {
		t.Fatalf("CountUnread failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 unread notifications, got %d", count)
	}
}
