package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// WebhookRepository handles webhook data access
type WebhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(ctx context.Context, webhook *models.Webhook) (*models.Webhook, error) {
	query := `
		INSERT INTO webhooks (project_id, name, url, secret, events, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, project_id, name, url, secret, events, is_active, created_by, created_at, updated_at
	`

	var created models.Webhook
	err := r.db.QueryRowContext(ctx, query,
		webhook.ProjectID,
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		pq.Array(webhook.Events),
		webhook.IsActive,
		webhook.CreatedBy,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.Name,
		&created.URL,
		&created.Secret,
		&created.Events,
		&created.IsActive,
		&created.CreatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves a webhook by ID
func (r *WebhookRepository) GetByID(ctx context.Context, id int) (*models.Webhook, error) {
	query := `
		SELECT id, project_id, name, url, secret, events, is_active, created_by, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`

	var webhook models.Webhook
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.ProjectID,
		&webhook.Name,
		&webhook.URL,
		&webhook.Secret,
		&webhook.Events,
		&webhook.IsActive,
		&webhook.CreatedBy,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// ListByProject retrieves all webhooks for a project
func (r *WebhookRepository) ListByProject(ctx context.Context, projectID int) ([]*models.Webhook, error) {
	query := `
		SELECT id, project_id, name, url, secret, events, is_active, created_by, created_at, updated_at
		FROM webhooks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	webhooks := make([]*models.Webhook, 0)
	for rows.Next() {
		var webhook models.Webhook
		err := rows.Scan(
			&webhook.ID,
			&webhook.ProjectID,
			&webhook.Name,
			&webhook.URL,
			&webhook.Secret,
			&webhook.Events,
			&webhook.IsActive,
			&webhook.CreatedBy,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, &webhook)
	}

	return webhooks, rows.Err()
}

// ListActiveByProjectAndEvent retrieves active webhooks for a project that subscribe to a specific event
func (r *WebhookRepository) ListActiveByProjectAndEvent(ctx context.Context, projectID int, eventType string) ([]*models.Webhook, error) {
	query := `
		SELECT id, project_id, name, url, secret, events, is_active, created_by, created_at, updated_at
		FROM webhooks
		WHERE project_id = $1 AND is_active = true AND $2 = ANY(events)
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	webhooks := make([]*models.Webhook, 0)
	for rows.Next() {
		var webhook models.Webhook
		err := rows.Scan(
			&webhook.ID,
			&webhook.ProjectID,
			&webhook.Name,
			&webhook.URL,
			&webhook.Secret,
			&webhook.Events,
			&webhook.IsActive,
			&webhook.CreatedBy,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, &webhook)
	}

	return webhooks, rows.Err()
}

// Update updates a webhook
func (r *WebhookRepository) Update(ctx context.Context, webhook *models.Webhook) error {
	query := `
		UPDATE webhooks
		SET name = $1, url = $2, secret = $3, events = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		pq.Array(webhook.Events),
		webhook.IsActive,
		webhook.ID,
	)

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

// Delete deletes a webhook
func (r *WebhookRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM webhooks WHERE id = $1`

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

// CreateDelivery creates a webhook delivery record
func (r *WebhookRepository) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) (*models.WebhookDelivery, error) {
	query := `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, response_status, response_body, error_message, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		delivery.WebhookID,
		delivery.EventType,
		delivery.Payload,
		delivery.ResponseStatus,
		delivery.ResponseBody,
		delivery.ErrorMessage,
		delivery.DeliveredAt,
	).Scan(&delivery.ID, &delivery.CreatedAt)

	if err != nil {
		return nil, err
	}

	return delivery, nil
}

// ListDeliveries retrieves recent deliveries for a webhook
func (r *WebhookRepository) ListDeliveries(ctx context.Context, webhookID int, limit int) ([]*models.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, error_message, delivered_at, created_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deliveries := make([]*models.WebhookDelivery, 0)
	for rows.Next() {
		var d models.WebhookDelivery
		err := rows.Scan(
			&d.ID,
			&d.WebhookID,
			&d.EventType,
			&d.Payload,
			&d.ResponseStatus,
			&d.ResponseBody,
			&d.ErrorMessage,
			&d.DeliveredAt,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, &d)
	}

	return deliveries, rows.Err()
}
