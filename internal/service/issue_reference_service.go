package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
)

// IssueReferenceService handles issue reference business logic
type IssueReferenceService struct {
	referenceRepo *repository.IssueReferenceRepository
	issueRepo     *repository.IssueRepository
	db            *sql.DB
}

// NewIssueReferenceService creates a new issue reference service
func NewIssueReferenceService(
	referenceRepo *repository.IssueReferenceRepository,
	issueRepo *repository.IssueRepository,
	db *sql.DB,
) *IssueReferenceService {
	return &IssueReferenceService{
		referenceRepo: referenceRepo,
		issueRepo:     issueRepo,
		db:            db,
	}
}

// ProcessReferences parses text for #issue_number references and creates reference records
// Only creates references to issues within the same project
// Returns the list of referenced issue IDs
func (s *IssueReferenceService) ProcessReferences(ctx context.Context, text string, sourceType string, sourceID int, projectID int) ([]int, error) {
	// Parse issue numbers from text
	issueNumbers := models.ParseIssueReferences(text)
	if len(issueNumbers) == 0 {
		return nil, nil
	}

	// Look up issues by number within the same project
	var referencedIssueIDs []int
	var references []*models.IssueReference

	for _, issueNumber := range issueNumbers {
		// Find issue by project_id and issue_number
		issue, err := s.issueRepo.GetByProjectAndNumber(ctx, projectID, issueNumber)
		if err != nil {
			if err == sql.ErrNoRows {
				// Issue doesn't exist in this project, skip
				continue
			}
			return nil, err
		}

		// Don't allow self-references (issue referencing itself)
		if sourceType == "issue" && issue.ID == sourceID {
			continue
		}

		referencedIssueIDs = append(referencedIssueIDs, issue.ID)

		reference := &models.IssueReference{
			SourceType:        sourceType,
			SourceID:          sourceID,
			ReferencedIssueID: issue.ID,
			CreatedAt:         time.Now(),
		}
		references = append(references, reference)
	}

	// Create reference records in batch
	if len(references) > 0 {
		err := s.referenceRepo.CreateBatch(ctx, references)
		if err != nil {
			return nil, err
		}
	}

	return referencedIssueIDs, nil
}

// GetReferencesFromSource retrieves all references made by a source (issue or comment)
func (s *IssueReferenceService) GetReferencesFromSource(ctx context.Context, sourceType string, sourceID int) ([]*models.IssueReference, error) {
	return s.referenceRepo.GetBySource(ctx, sourceType, sourceID)
}

// GetReferencesToIssue retrieves all references pointing to an issue
func (s *IssueReferenceService) GetReferencesToIssue(ctx context.Context, issueID int) ([]*models.IssueReference, error) {
	return s.referenceRepo.GetReferencesToIssue(ctx, issueID)
}
