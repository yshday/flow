package service

import (
	"context"
	"log"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/email"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	userRepo         *repository.UserRepository
	emailClient      *email.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	emailClient *email.Client,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		emailClient:      emailClient,
	}
}

// Create creates a new notification
func (s *NotificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, pkgerrors.ErrNotFound
	}

	// Verify actor exists if provided
	if req.ActorID != nil {
		_, err := s.userRepo.GetByID(ctx, *req.ActorID)
		if err != nil {
			return nil, pkgerrors.ErrNotFound
		}
	}

	notification := &models.Notification{
		UserID:     req.UserID,
		ActorID:    req.ActorID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Action:     req.Action,
		Title:      req.Title,
		Message:    req.Message,
		Read:       false,
	}

	return s.notificationRepo.Create(ctx, notification)
}

// GetByID retrieves a notification by ID (only if it belongs to the requesting user)
func (s *NotificationService) GetByID(ctx context.Context, id int, userID int) (*models.Notification, error) {
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if notification belongs to the user
	if notification.UserID != userID {
		return nil, pkgerrors.ErrForbidden
	}

	return notification, nil
}

// ListByUserID lists all notifications for a user with pagination
func (s *NotificationService) ListByUserID(ctx context.Context, userID int, limit int, offset int) ([]*models.Notification, error) {
	return s.notificationRepo.ListByUserID(ctx, userID, limit, offset)
}

// ListUnreadByUserID lists only unread notifications for a user with pagination
func (s *NotificationService) ListUnreadByUserID(ctx context.Context, userID int, limit int, offset int) ([]*models.Notification, error) {
	return s.notificationRepo.ListUnreadByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks specific notifications as read (only if they belong to the requesting user)
func (s *NotificationService) MarkAsRead(ctx context.Context, req *models.MarkAsReadRequest, userID int) error {
	if len(req.NotificationIDs) == 0 {
		return nil
	}

	// Verify all notifications belong to the user
	for _, notificationID := range req.NotificationIDs {
		notification, err := s.notificationRepo.GetByID(ctx, notificationID)
		if err != nil {
			return err
		}

		if notification.UserID != userID {
			return pkgerrors.ErrForbidden
		}
	}

	return s.notificationRepo.MarkAsRead(ctx, req.NotificationIDs)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID int) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// Delete deletes a notification (only if it belongs to the requesting user)
func (s *NotificationService) Delete(ctx context.Context, id int, userID int) error {
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if notification belongs to the user
	if notification.UserID != userID {
		return pkgerrors.ErrForbidden
	}

	return s.notificationRepo.Delete(ctx, id)
}

// CountUnread counts the number of unread notifications for a user
func (s *NotificationService) CountUnread(ctx context.Context, userID int) (int, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}

// CreateForIssueCreated creates a notification when an issue is created
func (s *NotificationService) CreateForIssueCreated(ctx context.Context, issueID int, issueKey, issueTitle string, actorID int, assigneeID *int, projectMemberIDs []int, projectName string) error {
	// Notify assignee if assigned
	if assigneeID != nil && *assigneeID != actorID {
		notification := &models.Notification{
			UserID:     *assigneeID,
			ActorID:    &actorID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   issueID,
			Action:     models.NotificationActionAssigned,
			Title:      "You were assigned to an issue",
			Message:    stringPtr("Issue: " + issueTitle),
			Read:       false,
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			return err
		}

		// Send email notification
		assignee, err := s.userRepo.GetByID(ctx, *assigneeID)
		if err == nil && assignee.Email != "" {
			actor, err := s.userRepo.GetByID(ctx, actorID)
			actorName := "Someone"
			if err == nil {
				actorName = actor.Username
			}

			err = s.emailClient.SendIssueAssigned(
				assignee.Email,
				issueKey,
				issueTitle,
				actorName,
				projectName,
			)
			if err != nil {
				log.Printf("Failed to send assignment email: %v", err)
			}
		}
	}

	// Notify project members (except actor and assignee)
	for _, memberID := range projectMemberIDs {
		if memberID == actorID || (assigneeID != nil && memberID == *assigneeID) {
			continue
		}

		notification := &models.Notification{
			UserID:     memberID,
			ActorID:    &actorID,
			EntityType: models.NotificationEntityIssue,
			EntityID:   issueID,
			Action:     models.NotificationActionCreated,
			Title:      "New issue created",
			Message:    stringPtr("Issue: " + issueTitle),
			Read:       false,
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateForComment creates a notification when a comment is added
func (s *NotificationService) CreateForComment(ctx context.Context, commentID int, issueID int, issueKey, issueTitle, commentText string, actorID int, issueCreatorID int, assigneeID *int, projectName string) error {
	// Get actor name
	actor, err := s.userRepo.GetByID(ctx, actorID)
	actorName := "Someone"
	if err == nil {
		actorName = actor.Username
	}

	// Notify issue creator if they're not the commenter
	if issueCreatorID != actorID {
		notification := &models.Notification{
			UserID:     issueCreatorID,
			ActorID:    &actorID,
			EntityType: models.NotificationEntityComment,
			EntityID:   commentID,
			Action:     models.NotificationActionCommented,
			Title:      "New comment on your issue",
			Message:    stringPtr("Issue: " + issueTitle),
			Read:       false,
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			return err
		}

		// Send email to issue creator
		issueCreator, err := s.userRepo.GetByID(ctx, issueCreatorID)
		if err == nil && issueCreator.Email != "" {
			err = s.emailClient.SendCommentAdded(
				issueCreator.Email,
				issueKey,
				issueTitle,
				actorName,
				commentText,
				projectName,
			)
			if err != nil {
				log.Printf("Failed to send comment email: %v", err)
			}
		}
	}

	// Notify assignee if they're not the commenter and not the creator
	if assigneeID != nil && *assigneeID != actorID && *assigneeID != issueCreatorID {
		notification := &models.Notification{
			UserID:     *assigneeID,
			ActorID:    &actorID,
			EntityType: models.NotificationEntityComment,
			EntityID:   commentID,
			Action:     models.NotificationActionCommented,
			Title:      "New comment on assigned issue",
			Message:    stringPtr("Issue: " + issueTitle),
			Read:       false,
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			return err
		}

		// Send email to assignee
		assignee, err := s.userRepo.GetByID(ctx, *assigneeID)
		if err == nil && assignee.Email != "" {
			err = s.emailClient.SendCommentAdded(
				assignee.Email,
				issueKey,
				issueTitle,
				actorName,
				commentText,
				projectName,
			)
			if err != nil {
				log.Printf("Failed to send comment email: %v", err)
			}
		}
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}
