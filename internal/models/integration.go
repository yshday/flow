package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// Integration types
const (
	IntegrationTypeSlack   = "slack"
	IntegrationTypeDiscord = "discord"
	IntegrationTypeTeams   = "teams"
	IntegrationTypeCustom  = "custom"
)

// AllIntegrationTypes returns all available integration types
func AllIntegrationTypes() []string {
	return []string{
		IntegrationTypeSlack,
		IntegrationTypeDiscord,
		IntegrationTypeTeams,
		IntegrationTypeCustom,
	}
}

// IntegrationSettings holds additional settings for an integration
type IntegrationSettings struct {
	Username  string `json:"username,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Color     string `json:"color,omitempty"` // Hex color for embeds
}

// Integration represents a messenger integration configuration
type Integration struct {
	ID         int                 `json:"id"`
	ProjectID  int                 `json:"project_id"`
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	WebhookURL string              `json:"-"` // Never expose webhook URL in JSON
	Channel    *string             `json:"channel,omitempty"`
	Events     pq.StringArray      `json:"events"`
	IsActive   bool                `json:"is_active"`
	Settings   IntegrationSettings `json:"settings"`
	CreatedBy  int                 `json:"created_by"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// IntegrationMessage represents a message delivery log
type IntegrationMessage struct {
	ID             int        `json:"id"`
	IntegrationID  int        `json:"integration_id"`
	EventType      string     `json:"event_type"`
	Message        string     `json:"message"`
	ResponseStatus *int       `json:"response_status"`
	ErrorMessage   *string    `json:"error_message"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateIntegrationRequest represents integration creation request
type CreateIntegrationRequest struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	WebhookURL string              `json:"webhook_url"`
	Channel    *string             `json:"channel,omitempty"`
	Events     []string            `json:"events"`
	Settings   IntegrationSettings `json:"settings,omitempty"`
}

// UpdateIntegrationRequest represents integration update request
type UpdateIntegrationRequest struct {
	Name       *string             `json:"name,omitempty"`
	WebhookURL *string             `json:"webhook_url,omitempty"`
	Channel    *string             `json:"channel,omitempty"`
	Events     []string            `json:"events,omitempty"`
	IsActive   *bool               `json:"is_active,omitempty"`
	Settings   *IntegrationSettings `json:"settings,omitempty"`
}

// SlackMessage represents a Slack incoming webhook message
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Blocks      []interface{}     `json:"blocks,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Fallback   string       `json:"fallback,omitempty"`
	Color      string       `json:"color,omitempty"`
	Pretext    string       `json:"pretext,omitempty"`
	AuthorName string       `json:"author_name,omitempty"`
	AuthorLink string       `json:"author_link,omitempty"`
	AuthorIcon string       `json:"author_icon,omitempty"`
	Title      string       `json:"title,omitempty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	Ts         int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// DiscordMessage represents a Discord webhook message
type DiscordMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed object
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"` // Decimal color
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Author      *DiscordEmbedAuthor `json:"author,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

// DiscordEmbedFooter represents a Discord embed footer
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedAuthor represents a Discord embed author
type DiscordEmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedField represents a Discord embed field
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// MarshalSettings converts IntegrationSettings to JSON for database storage
func (s *IntegrationSettings) MarshalJSON() ([]byte, error) {
	type Alias IntegrationSettings
	return json.Marshal((*Alias)(s))
}

// UnmarshalSettings parses JSON from database into IntegrationSettings
func (s *IntegrationSettings) UnmarshalJSON(data []byte) error {
	type Alias IntegrationSettings
	return json.Unmarshal(data, (*Alias)(s))
}
