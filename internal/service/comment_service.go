package service

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
	"github.com/yourusername/issue-tracker/pkg/markdown"
)

// CommentService handles comment business logic
type CommentService struct {
	commentRepo      *repository.CommentRepository
	issueRepo        *repository.IssueRepository
	authService      *AuthorizationService
	db               *sql.DB
	markdownRenderer *markdown.Renderer
	mentionService   *MentionService
	referenceService *IssueReferenceService
	webhookService   *WebhookService
}

// NewCommentService creates a new comment service
func NewCommentService(commentRepo *repository.CommentRepository, issueRepo *repository.IssueRepository, authService *AuthorizationService, db *sql.DB, mdRenderer *markdown.Renderer, mentionService *MentionService, referenceService *IssueReferenceService, webhookService *WebhookService) *CommentService {
	return &CommentService{
		commentRepo:      commentRepo,
		issueRepo:        issueRepo,
		authService:      authService,
		db:               db,
		markdownRenderer: mdRenderer,
		mentionService:   mentionService,
		referenceService: referenceService,
		webhookService:   webhookService,
	}
}

// Create creates a new comment
func (s *CommentService) Create(ctx context.Context, issueID int, req *models.CreateCommentRequest, userID int) (*models.Comment, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	comment := &models.Comment{
		IssueID: issueID,
		UserID:  userID,
		Content: req.Content,
	}

	// Render markdown content to HTML
	if req.Content != "" {
		html := s.markdownRenderer.RenderToHTML(req.Content)
		comment.ContentHTML = &html
	}

	created, err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Process @mentions and #issue_number references in comment content
	if req.Content != "" {
		mentionedUserIDs, err := s.mentionService.ProcessMentions(ctx, req.Content, "comment", created.ID, userID)
		if err != nil {
			// Log error but don't fail the whole operation
			// TODO: Add proper logging
		} else if len(mentionedUserIDs) > 0 {
			// Create notifications for mentioned users
			_ = s.mentionService.CreateNotificationsForMentions(ctx, mentionedUserIDs, "comment", created.ID, userID)
		}

		// Process #issue_number references
		_, _ = s.referenceService.ProcessReferences(ctx, req.Content, "comment", created.ID, issue.ProjectID)
	}

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventCommentCreated, userID, created)
	}

	return created, nil
}

// ListByIssueID lists all comments for an issue
func (s *CommentService) ListByIssueID(ctx context.Context, issueID int, userID int) ([]*models.Comment, error) {
	// Check if issue exists
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the project (viewers can read)
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return s.commentRepo.ListByIssueID(ctx, issueID)
}

// Update updates a comment
func (s *CommentService) Update(ctx context.Context, id int, req *models.UpdateCommentRequest, userID int) (*models.Comment, error) {
	// Get existing comment
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only comment author can update
	if comment.UserID != userID {
		return nil, pkgerrors.ErrForbidden
	}

	comment.Content = req.Content

	// Render markdown content to HTML
	if req.Content != "" {
		html := s.markdownRenderer.RenderToHTML(req.Content)
		comment.ContentHTML = &html
	} else {
		comment.ContentHTML = nil
	}

	// Get issue to know the project ID for references
	issue, err := s.issueRepo.GetByID(ctx, comment.IssueID)
	if err != nil {
		return nil, err
	}

	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Handle mentions and references when content is updated
	// Delete old mentions and references for this comment
	_ = s.mentionService.mentionRepo.DeleteByEntity(ctx, "comment", id)
	_ = s.referenceService.referenceRepo.DeleteBySource(ctx, "comment", id)

	// Process new mentions and references
	if req.Content != "" {
		mentionedUserIDs, err := s.mentionService.ProcessMentions(ctx, req.Content, "comment", id, userID)
		if err == nil && len(mentionedUserIDs) > 0 {
			// Create notifications for mentioned users
			_ = s.mentionService.CreateNotificationsForMentions(ctx, mentionedUserIDs, "comment", id, userID)
		}

		// Process #issue_number references
		_, _ = s.referenceService.ProcessReferences(ctx, req.Content, "comment", id, issue.ProjectID)
	}

	// Get updated comment
	updated, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventCommentUpdated, userID, updated)
	}

	return updated, nil
}

// Delete deletes a comment
func (s *CommentService) Delete(ctx context.Context, id int, userID int) error {
	// Get comment
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get issue to check project permissions
	issue, err := s.issueRepo.GetByID(ctx, comment.IssueID)
	if err != nil {
		return err
	}

	// Copy comment data before deletion for webhook
	deletedComment := *comment

	// Allow deletion if user is comment author
	if comment.UserID == userID {
		if err := s.commentRepo.Delete(ctx, id); err != nil {
			return err
		}
		// Deliver webhook event (use background context since this runs async)
		if s.webhookService != nil {
			go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventCommentDeleted, userID, &deletedComment)
		}
		return nil
	}

	// Otherwise, check if user has admin permission (only admins/owners can delete others' comments)
	if err := s.authService.CheckAdminPermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Deliver webhook event (use background context since this runs async)
	if s.webhookService != nil {
		go s.webhookService.DeliverEvent(context.Background(), issue.ProjectID, models.EventCommentDeleted, userID, &deletedComment)
	}

	return nil
}

// userHasAccess checks if user has any access to the project
func (s *CommentService) userHasAccess(ctx context.Context, userID int, projectID int) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// userHasPermission checks if user has specific role in the project
func (s *CommentService) userHasPermission(ctx context.Context, userID int, projectID int, allowedRoles []models.ProjectRole) (bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	for _, allowedRole := range allowedRoles {
		if models.ProjectRole(role) == allowedRole {
			return true, nil
		}
	}

	return false, nil
}
