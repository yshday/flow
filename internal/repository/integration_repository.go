package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/yourusername/issue-tracker/internal/models"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// IntegrationRepository handles integration data access
type IntegrationRepository struct {
	db *sql.DB
}

// NewIntegrationRepository creates a new integration repository
func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

// Create creates a new integration
func (r *IntegrationRepository) Create(ctx context.Context, integration *models.Integration) (*models.Integration, error) {
	settingsJSON, err := json.Marshal(integration.Settings)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO integrations (project_id, name, type, webhook_url, channel, events, is_active, settings, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, project_id, name, type, webhook_url, channel, events, is_active, settings, created_by, created_at, updated_at
	`

	var created models.Integration
	var settingsBytes []byte
	err = r.db.QueryRowContext(ctx, query,
		integration.ProjectID,
		integration.Name,
		integration.Type,
		integration.WebhookURL,
		integration.Channel,
		pq.Array(integration.Events),
		integration.IsActive,
		settingsJSON,
		integration.CreatedBy,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.Name,
		&created.Type,
		&created.WebhookURL,
		&created.Channel,
		&created.Events,
		&created.IsActive,
		&settingsBytes,
		&created.CreatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(settingsBytes, &created.Settings); err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByID retrieves an integration by ID
func (r *IntegrationRepository) GetByID(ctx context.Context, id int) (*models.Integration, error) {
	query := `
		SELECT id, project_id, name, type, webhook_url, channel, events, is_active, settings, created_by, created_at, updated_at
		FROM integrations
		WHERE id = $1
	`

	var integration models.Integration
	var settingsBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&integration.ID,
		&integration.ProjectID,
		&integration.Name,
		&integration.Type,
		&integration.WebhookURL,
		&integration.Channel,
		&integration.Events,
		&integration.IsActive,
		&settingsBytes,
		&integration.CreatedBy,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(settingsBytes, &integration.Settings); err != nil {
		return nil, err
	}

	return &integration, nil
}

// ListByProject retrieves all integrations for a project
func (r *IntegrationRepository) ListByProject(ctx context.Context, projectID int) ([]*models.Integration, error) {
	query := `
		SELECT id, project_id, name, type, webhook_url, channel, events, is_active, settings, created_by, created_at, updated_at
		FROM integrations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	integrations := make([]*models.Integration, 0)
	for rows.Next() {
		var integration models.Integration
		var settingsBytes []byte
		err := rows.Scan(
			&integration.ID,
			&integration.ProjectID,
			&integration.Name,
			&integration.Type,
			&integration.WebhookURL,
			&integration.Channel,
			&integration.Events,
			&integration.IsActive,
			&settingsBytes,
			&integration.CreatedBy,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(settingsBytes, &integration.Settings); err != nil {
			return nil, err
		}
		integrations = append(integrations, &integration)
	}

	return integrations, rows.Err()
}

// ListActiveByProjectAndEvent retrieves active integrations for a project that subscribe to a specific event
func (r *IntegrationRepository) ListActiveByProjectAndEvent(ctx context.Context, projectID int, eventType string) ([]*models.Integration, error) {
	query := `
		SELECT id, project_id, name, type, webhook_url, channel, events, is_active, settings, created_by, created_at, updated_at
		FROM integrations
		WHERE project_id = $1 AND is_active = true AND $2 = ANY(events)
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	integrations := make([]*models.Integration, 0)
	for rows.Next() {
		var integration models.Integration
		var settingsBytes []byte
		err := rows.Scan(
			&integration.ID,
			&integration.ProjectID,
			&integration.Name,
			&integration.Type,
			&integration.WebhookURL,
			&integration.Channel,
			&integration.Events,
			&integration.IsActive,
			&settingsBytes,
			&integration.CreatedBy,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(settingsBytes, &integration.Settings); err != nil {
			return nil, err
		}
		integrations = append(integrations, &integration)
	}

	return integrations, rows.Err()
}

// Update updates an integration
func (r *IntegrationRepository) Update(ctx context.Context, integration *models.Integration) error {
	settingsJSON, err := json.Marshal(integration.Settings)
	if err != nil {
		return err
	}

	query := `
		UPDATE integrations
		SET name = $1, webhook_url = $2, channel = $3, events = $4, is_active = $5, settings = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		integration.Name,
		integration.WebhookURL,
		integration.Channel,
		pq.Array(integration.Events),
		integration.IsActive,
		settingsJSON,
		integration.ID,
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

// Delete deletes an integration
func (r *IntegrationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM integrations WHERE id = $1`

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

// CreateMessage creates an integration message record
func (r *IntegrationRepository) CreateMessage(ctx context.Context, msg *models.IntegrationMessage) (*models.IntegrationMessage, error) {
	query := `
		INSERT INTO integration_messages (integration_id, event_type, message, response_status, error_message, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		msg.IntegrationID,
		msg.EventType,
		msg.Message,
		msg.ResponseStatus,
		msg.ErrorMessage,
		msg.DeliveredAt,
	).Scan(&msg.ID, &msg.CreatedAt)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

// ListMessages retrieves recent messages for an integration
func (r *IntegrationRepository) ListMessages(ctx context.Context, integrationID int, limit int) ([]*models.IntegrationMessage, error) {
	query := `
		SELECT id, integration_id, event_type, message, response_status, error_message, delivered_at, created_at
		FROM integration_messages
		WHERE integration_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, integrationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*models.IntegrationMessage, 0)
	for rows.Next() {
		var m models.IntegrationMessage
		err := rows.Scan(
			&m.ID,
			&m.IntegrationID,
			&m.EventType,
			&m.Message,
			&m.ResponseStatus,
			&m.ErrorMessage,
			&m.DeliveredAt,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}

	return messages, rows.Err()
}
