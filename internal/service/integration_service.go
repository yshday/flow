package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// IntegrationService handles integration business logic
type IntegrationService struct {
	integrationRepo *repository.IntegrationRepository
	authService     *AuthorizationService
	httpClient      *http.Client
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(integrationRepo *repository.IntegrationRepository, authService *AuthorizationService) *IntegrationService {
	return &IntegrationService{
		integrationRepo: integrationRepo,
		authService:     authService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Create creates a new integration
func (s *IntegrationService) Create(ctx context.Context, projectID int, req *models.CreateIntegrationRequest, userID int) (*models.Integration, error) {
	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Validate integration type
	if !s.isValidType(req.Type) {
		return nil, pkgerrors.ErrValidation
	}

	// Validate events
	if err := s.validateEvents(req.Events); err != nil {
		return nil, err
	}

	integration := &models.Integration{
		ProjectID:  projectID,
		Name:       req.Name,
		Type:       req.Type,
		WebhookURL: req.WebhookURL,
		Channel:    req.Channel,
		Events:     req.Events,
		IsActive:   true,
		Settings:   req.Settings,
		CreatedBy:  userID,
	}

	return s.integrationRepo.Create(ctx, integration)
}

// GetByID retrieves an integration by ID
func (s *IntegrationService) GetByID(ctx context.Context, id int, userID int) (*models.Integration, error) {
	integration, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check access
	if err := s.authService.CheckProjectAccess(ctx, integration.ProjectID, userID); err != nil {
		return nil, err
	}

	return integration, nil
}

// List retrieves all integrations for a project
func (s *IntegrationService) List(ctx context.Context, projectID int, userID int) ([]*models.Integration, error) {
	// Check access
	if err := s.authService.CheckProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	return s.integrationRepo.ListByProject(ctx, projectID)
}

// Update updates an integration
func (s *IntegrationService) Update(ctx context.Context, id int, req *models.UpdateIntegrationRequest, userID int) (*models.Integration, error) {
	integration, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, integration.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		integration.Name = *req.Name
	}
	if req.WebhookURL != nil {
		integration.WebhookURL = *req.WebhookURL
	}
	if req.Channel != nil {
		integration.Channel = req.Channel
	}
	if req.Events != nil {
		if err := s.validateEvents(req.Events); err != nil {
			return nil, err
		}
		integration.Events = req.Events
	}
	if req.IsActive != nil {
		integration.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		integration.Settings = *req.Settings
	}

	if err := s.integrationRepo.Update(ctx, integration); err != nil {
		return nil, err
	}

	return s.integrationRepo.GetByID(ctx, id)
}

// Delete deletes an integration
func (s *IntegrationService) Delete(ctx context.Context, id int, userID int) error {
	integration, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, integration.ProjectID, userID); err != nil {
		return err
	}

	return s.integrationRepo.Delete(ctx, id)
}

// GetMessages retrieves recent messages for an integration
func (s *IntegrationService) GetMessages(ctx context.Context, integrationID int, userID int, limit int) ([]*models.IntegrationMessage, error) {
	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	// Check access
	if err := s.authService.CheckProjectAccess(ctx, integration.ProjectID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.integrationRepo.ListMessages(ctx, integrationID, limit)
}

// SendEvent sends event notifications to all active integrations
func (s *IntegrationService) SendEvent(ctx context.Context, projectID int, eventType string, data interface{}) error {
	integrations, err := s.integrationRepo.ListActiveByProjectAndEvent(ctx, projectID, eventType)
	if err != nil {
		return err
	}

	if len(integrations) == 0 {
		return nil
	}

	// Send to each integration asynchronously
	for _, integration := range integrations {
		go s.sendToIntegration(context.Background(), integration, eventType, data)
	}

	return nil
}

// sendToIntegration sends a message to a single integration
func (s *IntegrationService) sendToIntegration(ctx context.Context, integration *models.Integration, eventType string, data interface{}) {
	var messageBytes []byte
	var err error

	// Format message based on integration type
	switch integration.Type {
	case models.IntegrationTypeSlack:
		messageBytes, err = s.formatSlackMessage(integration, eventType, data)
	case models.IntegrationTypeDiscord:
		messageBytes, err = s.formatDiscordMessage(integration, eventType, data)
	default:
		// For custom and other types, send a simple JSON payload
		messageBytes, err = s.formatGenericMessage(eventType, data)
	}

	msg := &models.IntegrationMessage{
		IntegrationID: integration.ID,
		EventType:     eventType,
		Message:       string(messageBytes),
	}

	if err != nil {
		errMsg := err.Error()
		msg.ErrorMessage = &errMsg
		s.integrationRepo.CreateMessage(ctx, msg)
		return
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", integration.WebhookURL, bytes.NewReader(messageBytes))
	if err != nil {
		errMsg := err.Error()
		msg.ErrorMessage = &errMsg
		s.integrationRepo.CreateMessage(ctx, msg)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	now := time.Now()
	msg.DeliveredAt = &now

	if err != nil {
		errMsg := err.Error()
		msg.ErrorMessage = &errMsg
		s.integrationRepo.CreateMessage(ctx, msg)
		return
	}
	defer resp.Body.Close()

	msg.ResponseStatus = &resp.StatusCode
	s.integrationRepo.CreateMessage(ctx, msg)
}

// formatSlackMessage formats a message for Slack
func (s *IntegrationService) formatSlackMessage(integration *models.Integration, eventType string, data interface{}) ([]byte, error) {
	title, description, color := s.getEventDetails(eventType, data)

	msg := models.SlackMessage{
		Username:  integration.Settings.Username,
		IconURL:   integration.Settings.IconURL,
		IconEmoji: integration.Settings.IconEmoji,
		Attachments: []models.SlackAttachment{
			{
				Color:    color,
				Title:    title,
				Text:     description,
				Footer:   "Flow Issue Tracker",
				Ts:       time.Now().Unix(),
			},
		},
	}

	if msg.Username == "" {
		msg.Username = "Flow"
	}

	return json.Marshal(msg)
}

// formatDiscordMessage formats a message for Discord
func (s *IntegrationService) formatDiscordMessage(integration *models.Integration, eventType string, data interface{}) ([]byte, error) {
	title, description, colorHex := s.getEventDetails(eventType, data)

	// Convert hex color to decimal
	colorDec := s.hexToDecimal(colorHex)

	msg := models.DiscordMessage{
		Username:  integration.Settings.Username,
		AvatarURL: integration.Settings.IconURL,
		Embeds: []models.DiscordEmbed{
			{
				Title:       title,
				Description: description,
				Color:       colorDec,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Footer: &models.DiscordEmbedFooter{
					Text: "Flow Issue Tracker",
				},
			},
		},
	}

	if msg.Username == "" {
		msg.Username = "Flow"
	}

	return json.Marshal(msg)
}

// formatGenericMessage formats a generic JSON message
func (s *IntegrationService) formatGenericMessage(eventType string, data interface{}) ([]byte, error) {
	payload := map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().UTC(),
		"data":      data,
	}
	return json.Marshal(payload)
}

// getEventDetails returns title, description, and color for an event
func (s *IntegrationService) getEventDetails(eventType string, data interface{}) (title, description, color string) {
	color = "#36a64f" // Default green

	switch eventType {
	case models.EventIssueCreated:
		if issue, ok := data.(*models.Issue); ok {
			title = fmt.Sprintf("New Issue: %s", issue.Title)
			description = fmt.Sprintf("Issue #%d was created", issue.IssueNumber)
			if issue.Description != nil && *issue.Description != "" {
				desc := *issue.Description
				if len(desc) > 200 {
					desc = desc[:200] + "..."
				}
				description += "\n" + desc
			}
		}
		color = "#36a64f" // Green

	case models.EventIssueUpdated:
		if issue, ok := data.(*models.Issue); ok {
			title = fmt.Sprintf("Issue Updated: %s", issue.Title)
			description = fmt.Sprintf("Issue #%d was updated", issue.IssueNumber)
		}
		color = "#2196F3" // Blue

	case models.EventIssueDeleted:
		if issue, ok := data.(*models.Issue); ok {
			title = fmt.Sprintf("Issue Deleted: %s", issue.Title)
			description = fmt.Sprintf("Issue #%d was deleted", issue.IssueNumber)
		}
		color = "#f44336" // Red

	case models.EventIssueMoved:
		if issue, ok := data.(*models.Issue); ok {
			title = fmt.Sprintf("Issue Moved: %s", issue.Title)
			description = fmt.Sprintf("Issue #%d was moved to a different column", issue.IssueNumber)
		}
		color = "#9C27B0" // Purple

	case models.EventCommentCreated:
		if comment, ok := data.(*models.Comment); ok {
			title = "New Comment"
			desc := comment.Content
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}
			description = desc
		}
		color = "#4CAF50" // Green

	case models.EventCommentUpdated:
		title = "Comment Updated"
		description = "A comment was updated"
		color = "#2196F3" // Blue

	case models.EventCommentDeleted:
		title = "Comment Deleted"
		description = "A comment was deleted"
		color = "#f44336" // Red

	default:
		title = strings.Replace(eventType, ".", " ", -1)
		title = strings.Title(title)
		description = fmt.Sprintf("Event: %s", eventType)
	}

	return title, description, color
}

// hexToDecimal converts a hex color string to decimal
func (s *IntegrationService) hexToDecimal(hex string) int {
	hex = strings.TrimPrefix(hex, "#")
	val, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return 0x36a64f // Default green
	}
	return int(val)
}

// isValidType checks if the integration type is valid
func (s *IntegrationService) isValidType(t string) bool {
	for _, valid := range models.AllIntegrationTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// validateEvents validates that all events are valid
func (s *IntegrationService) validateEvents(events []string) error {
	validEvents := make(map[string]bool)
	for _, e := range models.AllWebhookEvents() {
		validEvents[e] = true
	}

	for _, e := range events {
		if !validEvents[e] {
			return pkgerrors.ErrValidation
		}
	}

	return nil
}
