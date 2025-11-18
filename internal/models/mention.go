package models

import (
	"regexp"
	"time"
)

// Mention represents a user mention in an issue or comment
type Mention struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	MentionedByUserID int       `json:"mentioned_by_user_id"`
	EntityType        string    `json:"entity_type"` // "issue" or "comment"
	EntityID          int       `json:"entity_id"`
	CreatedAt         time.Time `json:"created_at"`

	// Related entities
	User        *User `json:"user,omitempty"`
	MentionedBy *User `json:"mentioned_by,omitempty"`
}

// ParseMentions extracts all @username mentions from text
// Returns a slice of unique usernames without the @ symbol
func ParseMentions(text string) []string {
	// Regex pattern to match @username (alphanumeric and underscore, 3-30 chars)
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_]{3,30})`)
	matches := mentionRegex.FindAllStringSubmatch(text, -1)

	// Use a map to track unique mentions
	uniqueMentions := make(map[string]bool)
	var mentions []string

	for _, match := range matches {
		if len(match) > 1 {
			username := match[1]
			if !uniqueMentions[username] {
				uniqueMentions[username] = true
				mentions = append(mentions, username)
			}
		}
	}

	return mentions
}
