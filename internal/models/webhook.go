package models

import (
	"time"

	"github.com/lib/pq"
)

// Webhook event types
const (
	EventIssueCreated   = "issue.created"
	EventIssueUpdated   = "issue.updated"
	EventIssueDeleted   = "issue.deleted"
	EventIssueMoved     = "issue.moved"
	EventCommentCreated = "comment.created"
	EventCommentUpdated = "comment.updated"
	EventCommentDeleted = "comment.deleted"
	EventLabelAdded     = "label.added"
	EventLabelRemoved   = "label.removed"
	EventTasklistItemCreated   = "tasklist_item.created"
	EventTasklistItemCompleted = "tasklist_item.completed"
)

// AllWebhookEvents returns all available webhook event types
func AllWebhookEvents() []string {
	return []string{
		EventIssueCreated,
		EventIssueUpdated,
		EventIssueDeleted,
		EventIssueMoved,
		EventCommentCreated,
		EventCommentUpdated,
		EventCommentDeleted,
		EventLabelAdded,
		EventLabelRemoved,
		EventTasklistItemCreated,
		EventTasklistItemCompleted,
	}
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID        int            `json:"id"`
	ProjectID int            `json:"project_id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Secret    *string        `json:"-"` // Never expose secret in JSON
	Events    pq.StringArray `json:"events"`
	IsActive  bool           `json:"is_active"`
	CreatedBy int            `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt log
type WebhookDelivery struct {
	ID             int        `json:"id"`
	WebhookID      int        `json:"webhook_id"`
	EventType      string     `json:"event_type"`
	Payload        string     `json:"payload"` // JSON string
	ResponseStatus *int       `json:"response_status"`
	ResponseBody   *string    `json:"response_body"`
	ErrorMessage   *string    `json:"error_message"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	ProjectID int         `json:"project_id"`
	Actor     *User       `json:"actor,omitempty"`
	Data      interface{} `json:"data"`
}

// CreateWebhookRequest represents webhook creation request
type CreateWebhookRequest struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Secret *string  `json:"secret,omitempty"`
	Events []string `json:"events"`
}

// UpdateWebhookRequest represents webhook update request
type UpdateWebhookRequest struct {
	Name     *string  `json:"name,omitempty"`
	URL      *string  `json:"url,omitempty"`
	Secret   *string  `json:"secret,omitempty"`
	Events   []string `json:"events,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
