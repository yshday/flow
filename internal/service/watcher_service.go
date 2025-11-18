package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// WatcherService handles issue watcher business logic
type WatcherService struct {
	watcherRepo      *repository.IssueWatcherRepository
	issueRepo        *repository.IssueRepository
	notificationRepo *repository.NotificationRepository
	authService      *AuthorizationService
	db               *sql.DB
}

// NewWatcherService creates a new watcher service
func NewWatcherService(
	watcherRepo *repository.IssueWatcherRepository,
	issueRepo *repository.IssueRepository,
	notificationRepo *repository.NotificationRepository,
	authService *AuthorizationService,
	db *sql.DB,
) *WatcherService {
	return &WatcherService{
		watcherRepo:      watcherRepo,
		issueRepo:        issueRepo,
		notificationRepo: notificationRepo,
		authService:      authService,
		db:               db,
	}
}

// WatchIssue subscribes a user to an issue
func (s *WatcherService) WatchIssue(ctx context.Context, userID int, projectID int, issueNumber int) error {
	// Get the issue
	issue, err := s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkgerrors.NewNotFoundError("issue not found")
		}
		return err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, userID, projectID); err != nil {
		return err
	}

	// Watch the issue
	return s.watcherRepo.Watch(ctx, userID, issue.ID)
}

// UnwatchIssue unsubscribes a user from an issue
func (s *WatcherService) UnwatchIssue(ctx context.Context, userID int, projectID int, issueNumber int) error {
	// Get the issue
	issue, err := s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkgerrors.NewNotFoundError("issue not found")
		}
		return err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, userID, projectID); err != nil {
		return err
	}

	// Unwatch the issue
	return s.watcherRepo.Unwatch(ctx, userID, issue.ID)
}

// IsWatchingIssue checks if a user is watching an issue
func (s *WatcherService) IsWatchingIssue(ctx context.Context, userID int, projectID int, issueNumber int) (bool, error) {
	// Get the issue
	issue, err := s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, pkgerrors.NewNotFoundError("issue not found")
		}
		return false, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, userID, projectID); err != nil {
		return false, err
	}

	return s.watcherRepo.IsWatching(ctx, userID, issue.ID)
}

// GetWatchersForIssue retrieves all watchers for an issue
func (s *WatcherService) GetWatchersForIssue(ctx context.Context, userID int, projectID int, issueNumber int) ([]*models.IssueWatcher, error) {
	// Get the issue
	issue, err := s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.NewNotFoundError("issue not found")
		}
		return nil, err
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	return s.watcherRepo.GetWatchersByIssue(ctx, issue.ID)
}

// GetWatchedIssuesForUser retrieves all issues a user is watching
func (s *WatcherService) GetWatchedIssuesForUser(ctx context.Context, userID int, limit, offset int) ([]*models.IssueWatcher, error) {
	// Users can only view their own watched issues
	return s.watcherRepo.GetWatchedIssuesByUser(ctx, userID, limit, offset)
}

// NotifyWatchers sends notifications to all watchers of an issue
func (s *WatcherService) NotifyWatchers(ctx context.Context, issueID int, action models.NotificationAction, actorID int, message string) error {
	// Get all watcher user IDs
	watcherUserIDs, err := s.watcherRepo.GetWatcherUserIDs(ctx, issueID)
	if err != nil {
		return err
	}

	// Don't notify the actor
	var notifyUserIDs []int
	for _, uid := range watcherUserIDs {
		if uid != actorID {
			notifyUserIDs = append(notifyUserIDs, uid)
		}
	}

	if len(notifyUserIDs) == 0 {
		return nil
	}

	// Create notifications for all watchers
	now := time.Now()

	for _, uid := range notifyUserIDs {
		notification := &models.Notification{
			UserID:     uid,
			ActorID:    &actorID,
			EntityType: "issue",
			EntityID:   issueID,
			Action:     action,
			Title:      getWatcherNotificationTitle(action),
			Message:    &message,
			Read:       false,
			CreatedAt:  now,
		}

		_, err := s.notificationRepo.Create(ctx, notification)
		if err != nil {
			// Log error but continue creating other notifications
			continue
		}
	}

	return nil
}

// getWatcherNotificationTitle returns the notification title based on action
func getWatcherNotificationTitle(action models.NotificationAction) string {
	switch action {
	case models.NotificationActionUpdated:
		return "Issue updated"
	case models.NotificationActionCommented:
		return "New comment on watched issue"
	case models.NotificationActionAssigned:
		return "Issue assigned"
	default:
		return "Issue activity"
	}
}
