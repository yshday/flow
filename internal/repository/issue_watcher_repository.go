package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/issue-tracker/internal/models"
)

// IssueWatcherRepository handles issue watcher data access
type IssueWatcherRepository struct {
	db *sql.DB
}

// NewIssueWatcherRepository creates a new issue watcher repository
func NewIssueWatcherRepository(db *sql.DB) *IssueWatcherRepository {
	return &IssueWatcherRepository{db: db}
}

// Watch adds a user as a watcher of an issue
func (r *IssueWatcherRepository) Watch(ctx context.Context, userID, issueID int) error {
	query := `
		INSERT INTO issue_watchers (user_id, issue_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, issue_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, userID, issueID)
	return err
}

// Unwatch removes a user as a watcher of an issue
func (r *IssueWatcherRepository) Unwatch(ctx context.Context, userID, issueID int) error {
	query := `DELETE FROM issue_watchers WHERE user_id = $1 AND issue_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, issueID)
	return err
}

// IsWatching checks if a user is watching an issue
func (r *IssueWatcherRepository) IsWatching(ctx context.Context, userID, issueID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM issue_watchers WHERE user_id = $1 AND issue_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, issueID).Scan(&exists)
	return exists, err
}

// GetWatchersByIssue retrieves all watchers for an issue
func (r *IssueWatcherRepository) GetWatchersByIssue(ctx context.Context, issueID int) ([]*models.IssueWatcher, error) {
	query := `
		SELECT iw.id, iw.user_id, iw.issue_id, iw.created_at,
		       u.id, u.username, u.email, u.avatar_url
		FROM issue_watchers iw
		LEFT JOIN users u ON iw.user_id = u.id
		WHERE iw.issue_id = $1
		ORDER BY iw.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	watchers := make([]*models.IssueWatcher, 0)
	for rows.Next() {
		watcher := &models.IssueWatcher{User: &models.User{}}
		err := rows.Scan(
			&watcher.ID,
			&watcher.UserID,
			&watcher.IssueID,
			&watcher.CreatedAt,
			&watcher.User.ID,
			&watcher.User.Username,
			&watcher.User.Email,
			&watcher.User.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}

	return watchers, rows.Err()
}

// GetWatchedIssuesByUser retrieves all issues a user is watching
func (r *IssueWatcherRepository) GetWatchedIssuesByUser(ctx context.Context, userID int, limit, offset int) ([]*models.IssueWatcher, error) {
	query := `
		SELECT iw.id, iw.user_id, iw.issue_id, iw.created_at,
		       i.id, i.project_id, i.issue_number, i.title, i.status, i.priority
		FROM issue_watchers iw
		LEFT JOIN issues i ON iw.issue_id = i.id
		WHERE iw.user_id = $1
		ORDER BY iw.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	watchers := make([]*models.IssueWatcher, 0)
	for rows.Next() {
		watcher := &models.IssueWatcher{Issue: &models.Issue{}}
		err := rows.Scan(
			&watcher.ID,
			&watcher.UserID,
			&watcher.IssueID,
			&watcher.CreatedAt,
			&watcher.Issue.ID,
			&watcher.Issue.ProjectID,
			&watcher.Issue.IssueNumber,
			&watcher.Issue.Title,
			&watcher.Issue.Status,
			&watcher.Issue.Priority,
		)
		if err != nil {
			return nil, err
		}
		watchers = append(watchers, watcher)
	}

	return watchers, rows.Err()
}

// GetWatcherUserIDs retrieves all user IDs watching an issue
func (r *IssueWatcherRepository) GetWatcherUserIDs(ctx context.Context, issueID int) ([]int, error) {
	query := `SELECT user_id FROM issue_watchers WHERE issue_id = $1`
	rows, err := r.db.QueryContext(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}
