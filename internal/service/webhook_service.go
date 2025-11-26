package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// WebhookService handles webhook business logic
type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	authService *AuthorizationService
	httpClient  *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(webhookRepo *repository.WebhookRepository, authService *AuthorizationService) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
		authService: authService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Create creates a new webhook
func (s *WebhookService) Create(ctx context.Context, projectID int, req *models.CreateWebhookRequest, userID int) (*models.Webhook, error) {
	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, projectID, userID); err != nil {
		return nil, err
	}

	// Validate events
	if err := s.validateEvents(req.Events); err != nil {
		return nil, err
	}

	webhook := &models.Webhook{
		ProjectID: projectID,
		Name:      req.Name,
		URL:       req.URL,
		Secret:    req.Secret,
		Events:    req.Events,
		IsActive:  true,
		CreatedBy: userID,
	}

	return s.webhookRepo.Create(ctx, webhook)
}

// GetByID retrieves a webhook by ID
func (s *WebhookService) GetByID(ctx context.Context, id int, userID int) (*models.Webhook, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check access
	if err := s.authService.CheckProjectAccess(ctx, webhook.ProjectID, userID); err != nil {
		return nil, err
	}

	return webhook, nil
}

// List retrieves all webhooks for a project
func (s *WebhookService) List(ctx context.Context, projectID int, userID int) ([]*models.Webhook, error) {
	// Check access
	if err := s.authService.CheckProjectAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	return s.webhookRepo.ListByProject(ctx, projectID)
}

// Update updates a webhook
func (s *WebhookService) Update(ctx context.Context, id int, req *models.UpdateWebhookRequest, userID int) (*models.Webhook, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, webhook.ProjectID, userID); err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Secret != nil {
		webhook.Secret = req.Secret
	}
	if req.Events != nil {
		if err := s.validateEvents(req.Events); err != nil {
			return nil, err
		}
		webhook.Events = req.Events
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return s.webhookRepo.GetByID(ctx, id)
}

// Delete deletes a webhook
func (s *WebhookService) Delete(ctx context.Context, id int, userID int) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check admin permission
	if err := s.authService.CheckAdminPermission(ctx, webhook.ProjectID, userID); err != nil {
		return err
	}

	return s.webhookRepo.Delete(ctx, id)
}

// GetDeliveries retrieves recent deliveries for a webhook
func (s *WebhookService) GetDeliveries(ctx context.Context, webhookID int, userID int, limit int) ([]*models.WebhookDelivery, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	// Check access
	if err := s.authService.CheckProjectAccess(ctx, webhook.ProjectID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.webhookRepo.ListDeliveries(ctx, webhookID, limit)
}

// DeliverEvent sends a webhook payload to all subscribers
func (s *WebhookService) DeliverEvent(ctx context.Context, projectID int, eventType string, actorID int, data interface{}) error {
	webhooks, err := s.webhookRepo.ListActiveByProjectAndEvent(ctx, projectID, eventType)
	if err != nil {
		return err
	}

	if len(webhooks) == 0 {
		return nil
	}

	// Create minimal actor info for the payload
	actor := &models.User{ID: actorID}

	payload := &models.WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now().UTC(),
		ProjectID: projectID,
		Actor:     actor,
		Data:      data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Deliver to each webhook asynchronously
	for _, webhook := range webhooks {
		go s.deliverToWebhook(context.Background(), webhook, eventType, payloadBytes)
	}

	return nil
}

// deliverToWebhook sends the payload to a single webhook
func (s *WebhookService) deliverToWebhook(ctx context.Context, webhook *models.Webhook, eventType string, payloadBytes []byte) {
	delivery := &models.WebhookDelivery{
		WebhookID: webhook.ID,
		EventType: eventType,
		Payload:   string(payloadBytes),
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		errMsg := err.Error()
		delivery.ErrorMessage = &errMsg
		s.webhookRepo.CreateDelivery(ctx, delivery)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", eventType)

	// Add HMAC signature if secret is set
	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := s.generateSignature(payloadBytes, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	}

	resp, err := s.httpClient.Do(req)
	now := time.Now()
	delivery.DeliveredAt = &now

	if err != nil {
		errMsg := err.Error()
		delivery.ErrorMessage = &errMsg
		s.webhookRepo.CreateDelivery(ctx, delivery)
		return
	}
	defer resp.Body.Close()

	delivery.ResponseStatus = &resp.StatusCode

	// Log the delivery
	s.webhookRepo.CreateDelivery(ctx, delivery)
}

// generateSignature creates an HMAC-SHA256 signature
func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// validateEvents validates that all events are valid
func (s *WebhookService) validateEvents(events []string) error {
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
