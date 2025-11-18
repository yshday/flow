package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/errors"
)

// ReactionService handles reaction business logic
type ReactionService struct {
	reactionRepo    *repository.ReactionRepository
	issueRepo       *repository.IssueRepository
	commentRepo     *repository.CommentRepository
	authzService    *AuthorizationService
	db              *sql.DB
}

// NewReactionService creates a new ReactionService
func NewReactionService(
	reactionRepo *repository.ReactionRepository,
	issueRepo *repository.IssueRepository,
	commentRepo *repository.CommentRepository,
	authzService *AuthorizationService,
	db *sql.DB,
) *ReactionService {
	return &ReactionService{
		reactionRepo: reactionRepo,
		issueRepo:    issueRepo,
		commentRepo:  commentRepo,
		authzService: authzService,
		db:           db,
	}
}

// AddReaction adds or toggles a reaction
func (s *ReactionService) AddReaction(ctx context.Context, userID int, entityType string, entityID int, emoji string) (*models.Reaction, error) {
	// Validate emoji
	if !models.IsValidEmoji(emoji) {
		return nil, errors.NewValidationError("invalid emoji type")
	}

	// Validate entity type
	if entityType != "issue" && entityType != "comment" {
		return nil, errors.NewValidationError("entity_type must be 'issue' or 'comment'")
	}

	// Check if entity exists and user has access
	var projectID int
	var err error

	if entityType == "issue" {
		issue, err := s.issueRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("issue not found")
			}
			return nil, err
		}
		projectID = issue.ProjectID

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, projectID, userID); err != nil {
			return nil, err
		}
	} else if entityType == "comment" {
		comment, err := s.commentRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("comment not found")
			}
			return nil, err
		}

		// Get issue to check project access
		issue, err := s.issueRepo.GetByID(ctx, comment.IssueID)
		if err != nil {
			return nil, err
		}
		projectID = issue.ProjectID

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, projectID, userID); err != nil {
			return nil, err
		}
	}

	// Check if reaction already exists
	existing, err := s.reactionRepo.GetByUser(userID, entityType, entityID, emoji)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Reaction exists, remove it (toggle off)
		err = s.reactionRepo.Delete(userID, entityType, entityID, emoji)
		if err != nil {
			return nil, err
		}
		return nil, nil // Return nil to indicate removal
	}

	// Create new reaction
	reaction := &models.Reaction{
		UserID:     userID,
		EntityType: entityType,
		EntityID:   entityID,
		Emoji:      emoji,
		CreatedAt:  time.Now(),
	}

	err = s.reactionRepo.Create(reaction)
	if err != nil {
		return nil, err
	}

	return reaction, nil
}

// RemoveReaction removes a reaction
func (s *ReactionService) RemoveReaction(ctx context.Context, userID int, entityType string, entityID int, emoji string) error {
	// Validate entity type
	if entityType != "issue" && entityType != "comment" {
		return errors.NewValidationError("entity_type must be 'issue' or 'comment'")
	}

	// Check if reaction exists
	existing, err := s.reactionRepo.GetByUser(userID, entityType, entityID, emoji)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.NewNotFoundError("reaction not found")
	}

	return s.reactionRepo.Delete(userID, entityType, entityID, emoji)
}

// GetReactions retrieves all reactions for an entity
func (s *ReactionService) GetReactions(ctx context.Context, userID int, entityType string, entityID int) ([]*models.Reaction, error) {
	// Validate entity type
	if entityType != "issue" && entityType != "comment" {
		return nil, errors.NewValidationError("entity_type must be 'issue' or 'comment'")
	}

	// Check if entity exists and user has access
	if entityType == "issue" {
		issue, err := s.issueRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("issue not found")
			}
			return nil, err
		}

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
			return nil, err
		}
	} else if entityType == "comment" {
		comment, err := s.commentRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("comment not found")
			}
			return nil, err
		}

		// Get issue to check project access
		issue, err := s.issueRepo.GetByID(ctx, comment.IssueID)
		if err != nil {
			return nil, err
		}

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetByEntity(entityType, entityID)
}

// GetReactionSummary retrieves aggregated reaction counts for an entity
func (s *ReactionService) GetReactionSummary(ctx context.Context, userID int, entityType string, entityID int) (*models.ReactionSummary, error) {
	// Validate entity type
	if entityType != "issue" && entityType != "comment" {
		return nil, errors.NewValidationError("entity_type must be 'issue' or 'comment'")
	}

	// Check if entity exists and user has access
	if entityType == "issue" {
		issue, err := s.issueRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("issue not found")
			}
			return nil, err
		}

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
			return nil, err
		}
	} else if entityType == "comment" {
		comment, err := s.commentRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewNotFoundError("comment not found")
			}
			return nil, err
		}

		// Get issue to check project access
		issue, err := s.issueRepo.GetByID(ctx, comment.IssueID)
		if err != nil {
			return nil, err
		}

		// Check if user has access to the project
		if err := s.authzService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetSummary(entityType, entityID)
}
