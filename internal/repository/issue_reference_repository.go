package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
)

// IssueReferenceRepository handles issue reference data access
type IssueReferenceRepository struct {
	db *sql.DB
}

// NewIssueReferenceRepository creates a new issue reference repository
func NewIssueReferenceRepository(db *sql.DB) *IssueReferenceRepository {
	return &IssueReferenceRepository{db: db}
}

// Create creates a new issue reference
func (r *IssueReferenceRepository) Create(ctx context.Context, ref *models.IssueReference) error {
	query := `
		INSERT INTO issue_references (source_type, source_id, referenced_issue_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_type, source_id, referenced_issue_id) DO NOTHING
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		ref.SourceType,
		ref.SourceID,
		ref.ReferencedIssueID,
		ref.CreatedAt,
	).Scan(&ref.ID)

	if err == sql.ErrNoRows {
		// Reference already exists, not an error
		return nil
	}

	return err
}

// CreateBatch creates multiple issue references in a single transaction
func (r *IssueReferenceRepository) CreateBatch(ctx context.Context, refs []*models.IssueReference) error {
	if len(refs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO issue_references (source_type, source_id, referenced_issue_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_type, source_id, referenced_issue_id) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ref := range refs {
		_, err = stmt.ExecContext(ctx,
			ref.SourceType,
			ref.SourceID,
			ref.ReferencedIssueID,
			ref.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBySource retrieves all references from a specific source (issue or comment)
func (r *IssueReferenceRepository) GetBySource(ctx context.Context, sourceType string, sourceID int) ([]*models.IssueReference, error) {
	query := `
		SELECT ir.id, ir.source_type, ir.source_id, ir.referenced_issue_id, ir.created_at,
		       i.id, i.project_id, i.title, i.status, i.priority, i.issue_number
		FROM issue_references ir
		LEFT JOIN issues i ON ir.referenced_issue_id = i.id
		WHERE ir.source_type = $1 AND ir.source_id = $2
		ORDER BY ir.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, sourceType, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := make([]*models.IssueReference, 0)
	for rows.Next() {
		ref := &models.IssueReference{ReferencedIssue: &models.Issue{}}
		err := rows.Scan(
			&ref.ID,
			&ref.SourceType,
			&ref.SourceID,
			&ref.ReferencedIssueID,
			&ref.CreatedAt,
			&ref.ReferencedIssue.ID,
			&ref.ReferencedIssue.ProjectID,
			&ref.ReferencedIssue.Title,
			&ref.ReferencedIssue.Status,
			&ref.ReferencedIssue.Priority,
			&ref.ReferencedIssue.IssueNumber,
		)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, rows.Err()
}

// GetReferencesToIssue retrieves all references pointing to a specific issue
func (r *IssueReferenceRepository) GetReferencesToIssue(ctx context.Context, issueID int) ([]*models.IssueReference, error) {
	query := `
		SELECT id, source_type, source_id, referenced_issue_id, created_at
		FROM issue_references
		WHERE referenced_issue_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := make([]*models.IssueReference, 0)
	for rows.Next() {
		ref := &models.IssueReference{}
		err := rows.Scan(
			&ref.ID,
			&ref.SourceType,
			&ref.SourceID,
			&ref.ReferencedIssueID,
			&ref.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, rows.Err()
}

// DeleteBySource deletes all references from a specific source
func (r *IssueReferenceRepository) DeleteBySource(ctx context.Context, sourceType string, sourceID int) error {
	query := `DELETE FROM issue_references WHERE source_type = $1 AND source_id = $2`
	_, err := r.db.ExecContext(ctx, query, sourceType, sourceID)
	return err
}
