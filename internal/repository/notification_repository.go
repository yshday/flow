package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// NotificationRepository handles notification data operations
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, actor_id, entity_type, entity_id, action, title, message, read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, user_id, actor_id, entity_type, entity_id, action, title, message, read, read_at, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		notification.UserID,
		notification.ActorID,
		notification.EntityType,
		notification.EntityID,
		notification.Action,
		notification.Title,
		notification.Message,
		notification.Read,
	).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ActorID,
		&notification.EntityType,
		&notification.EntityID,
		&notification.Action,
		&notification.Title,
		&notification.Message,
		&notification.Read,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return notification, nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id int) (*models.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, entity_type, entity_id, action, title, message, read, read_at, created_at
		FROM notifications
		WHERE id = $1
	`

	notification := &models.Notification{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ActorID,
		&notification.EntityType,
		&notification.EntityID,
		&notification.Action,
		&notification.Title,
		&notification.Message,
		&notification.Read,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return notification, nil
}

// ListByUserID lists all notifications for a user with pagination
func (r *NotificationRepository) ListByUserID(ctx context.Context, userID int, limit int, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, entity_type, entity_id, action, title, message, read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		notification := &models.Notification{}
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ActorID,
			&notification.EntityType,
			&notification.EntityID,
			&notification.Action,
			&notification.Title,
			&notification.Message,
			&notification.Read,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

// ListUnreadByUserID lists only unread notifications for a user with pagination
func (r *NotificationRepository) ListUnreadByUserID(ctx context.Context, userID int, limit int, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, entity_type, entity_id, action, title, message, read, read_at, created_at
		FROM notifications
		WHERE user_id = $1 AND read = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		notification := &models.Notification{}
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ActorID,
			&notification.EntityType,
			&notification.EntityID,
			&notification.Action,
			&notification.Title,
			&notification.Message,
			&notification.Read,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

// Update updates a notification
func (r *NotificationRepository) Update(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	query := `
		UPDATE notifications
		SET read = $1, read_at = $2
		WHERE id = $3
		RETURNING id, user_id, actor_id, entity_type, entity_id, action, title, message, read, read_at, created_at
	`

	err := r.db.QueryRowContext(ctx, query, notification.Read, notification.ReadAt, notification.ID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ActorID,
		&notification.EntityType,
		&notification.EntityID,
		&notification.Action,
		&notification.Title,
		&notification.Message,
		&notification.Read,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return notification, nil
}

// MarkAsRead marks specific notifications as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationIDs []int) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	query := `
		UPDATE notifications
		SET read = true, read_at = $1
		WHERE id = ANY($2)
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), pq.Array(notificationIDs))
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID int) error {
	query := `
		UPDATE notifications
		SET read = true, read_at = $1
		WHERE user_id = $2 AND read = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}

// CountUnread counts the number of unread notifications for a user
func (r *NotificationRepository) CountUnread(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read = false
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
