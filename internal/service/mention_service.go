package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
)

// MentionService handles mention business logic
type MentionService struct {
	mentionRepo      *repository.MentionRepository
	userRepo         *repository.UserRepository
	notificationRepo *repository.NotificationRepository
	db               *sql.DB
}

// NewMentionService creates a new mention service
func NewMentionService(
	mentionRepo *repository.MentionRepository,
	userRepo *repository.UserRepository,
	notificationRepo *repository.NotificationRepository,
	db *sql.DB,
) *MentionService {
	return &MentionService{
		mentionRepo:      mentionRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		db:               db,
	}
}

// ProcessMentions parses text for @mentions and creates mention records
// Returns the list of mentioned user IDs
func (s *MentionService) ProcessMentions(ctx context.Context, text string, entityType string, entityID int, mentionedByUserID int) ([]int, error) {
	// Parse usernames from text
	usernames := models.ParseMentions(text)
	if len(usernames) == 0 {
		return nil, nil
	}

	// Look up users by username
	var mentionedUserIDs []int
	var mentions []*models.Mention

	for _, username := range usernames {
		user, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			if err == sql.ErrNoRows {
				// User doesn't exist, skip
				continue
			}
			return nil, err
		}

		// Don't allow self-mentions
		if user.ID == mentionedByUserID {
			continue
		}

		mentionedUserIDs = append(mentionedUserIDs, user.ID)

		mention := &models.Mention{
			UserID:            user.ID,
			MentionedByUserID: mentionedByUserID,
			EntityType:        entityType,
			EntityID:          entityID,
			CreatedAt:         time.Now(),
		}
		mentions = append(mentions, mention)
	}

	// Create mention records in batch
	if len(mentions) > 0 {
		err := s.mentionRepo.CreateBatch(ctx, mentions)
		if err != nil {
			return nil, err
		}
	}

	return mentionedUserIDs, nil
}

// CreateNotificationsForMentions creates notifications for all mentioned users
func (s *MentionService) CreateNotificationsForMentions(ctx context.Context, mentionedUserIDs []int, entityType string, entityID int, mentionedByUserID int) error {
	if len(mentionedUserIDs) == 0 {
		return nil
	}

	// Determine entity type for notification
	var notifEntityType models.NotificationEntityType
	if entityType == "issue" {
		notifEntityType = models.NotificationEntityIssue
	} else {
		notifEntityType = models.NotificationEntityComment
	}

	// Create a notification for each mentioned user
	for _, userID := range mentionedUserIDs {
		message := getMentionNotificationMessage(entityType)
		notification := &models.Notification{
			UserID:     userID,
			ActorID:    &mentionedByUserID,
			EntityType: notifEntityType,
			EntityID:   entityID,
			Action:     models.NotificationActionMentioned,
			Title:      "You were mentioned",
			Message:    &message,
			Read:       false,
			CreatedAt:  time.Now(),
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			// Log error but don't fail the whole operation
			continue
		}
	}

	return nil
}

// GetMentionsForUser retrieves all mentions for a user with pagination
func (s *MentionService) GetMentionsForUser(ctx context.Context, userID int, limit, offset int) ([]*models.Mention, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.mentionRepo.GetByUser(ctx, userID, limit, offset)
}

// getMentionNotificationMessage returns a user-friendly message based on entity type
func getMentionNotificationMessage(entityType string) string {
	switch entityType {
	case "issue":
		return "You were mentioned in an issue"
	case "comment":
		return "You were mentioned in a comment"
	default:
		return "You were mentioned"
	}
}
