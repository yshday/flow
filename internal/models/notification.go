package models

import (
	"time"
)

// NotificationEntityType represents the type of entity a notification is about
type NotificationEntityType string

const (
	NotificationEntityIssue   NotificationEntityType = "issue"
	NotificationEntityComment NotificationEntityType = "comment"
	NotificationEntityProject NotificationEntityType = "project"
)

// NotificationAction represents the action that triggered the notification
type NotificationAction string

const (
	NotificationActionCreated  NotificationAction = "created"
	NotificationActionUpdated  NotificationAction = "updated"
	NotificationActionDeleted  NotificationAction = "deleted"
	NotificationActionAssigned NotificationAction = "assigned"
	NotificationActionMentioned NotificationAction = "mentioned"
	NotificationActionCommented NotificationAction = "commented"
)

// Notification represents a user notification
type Notification struct {
	ID         int                    `json:"id"`
	UserID     int                    `json:"user_id"`
	ActorID    *int                   `json:"actor_id,omitempty"`
	EntityType NotificationEntityType `json:"entity_type"`
	EntityID   int                    `json:"entity_id"`
	Action     NotificationAction     `json:"action"`
	Title      string                 `json:"title"`
	Message    *string                `json:"message,omitempty"`
	Read       bool                   `json:"read"`
	ReadAt     *time.Time             `json:"read_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`

	// Related entities (optional, for enriched responses)
	Actor *User `json:"actor,omitempty"`
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID     int                    `json:"user_id" validate:"required"`
	ActorID    *int                   `json:"actor_id,omitempty"`
	EntityType NotificationEntityType `json:"entity_type" validate:"required,oneof=issue comment project"`
	EntityID   int                    `json:"entity_id" validate:"required"`
	Action     NotificationAction     `json:"action" validate:"required,oneof=created updated deleted assigned mentioned commented"`
	Title      string                 `json:"title" validate:"required,min=1,max=255"`
	Message    *string                `json:"message,omitempty"`
}

// MarkAsReadRequest represents the request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []int `json:"notification_ids" validate:"required,min=1"`
}
