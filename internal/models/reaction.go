package models

import "time"

// Reaction represents an emoji reaction on an issue or comment
type Reaction struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	EntityType string    `json:"entity_type"` // "issue" or "comment"
	EntityID   int       `json:"entity_id"`
	Emoji      string    `json:"emoji"`
	CreatedAt  time.Time `json:"created_at"`
	User       *User     `json:"user,omitempty"`
}

// ValidEmojis lists all supported emoji types
var ValidEmojis = []string{
	"thumbs_up",
	"thumbs_down",
	"laugh",
	"hooray",
	"confused",
	"heart",
	"rocket",
	"eyes",
}

// IsValidEmoji checks if the emoji is supported
func IsValidEmoji(emoji string) bool {
	for _, valid := range ValidEmojis {
		if emoji == valid {
			return true
		}
	}
	return false
}

// ReactionSummary represents aggregated reaction counts for an entity
type ReactionSummary struct {
	EntityType string                 `json:"entity_type"`
	EntityID   int                    `json:"entity_id"`
	Reactions  map[string]int         `json:"reactions"`          // emoji -> count
	UserReactions map[string][]int    `json:"user_reactions"`     // emoji -> list of user IDs
}
