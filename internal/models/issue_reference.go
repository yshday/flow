package models

import (
	"regexp"
	"strconv"
	"time"
)

// IssueReference represents a reference from an issue or comment to another issue
type IssueReference struct {
	ID                int       `json:"id"`
	SourceType        string    `json:"source_type"` // "issue" or "comment"
	SourceID          int       `json:"source_id"`
	ReferencedIssueID int       `json:"referenced_issue_id"`
	CreatedAt         time.Time `json:"created_at"`

	// Related entities
	ReferencedIssue *Issue `json:"referenced_issue,omitempty"`
}

// ParseIssueReferences extracts all #issue_number references from text
// Returns a slice of unique issue numbers (as integers)
func ParseIssueReferences(text string) []int {
	// Regex pattern to match #123 (issue number references)
	// Matches # followed by 1-9 digits
	referenceRegex := regexp.MustCompile(`#(\d{1,9})\b`)
	matches := referenceRegex.FindAllStringSubmatch(text, -1)

	// Use a map to track unique issue numbers
	uniqueNumbers := make(map[int]bool)
	var issueNumbers []int

	for _, match := range matches {
		if len(match) > 1 {
			// Convert string to int
			issueNumber, err := strconv.Atoi(match[1])
			if err != nil || issueNumber <= 0 {
				continue
			}

			if !uniqueNumbers[issueNumber] {
				uniqueNumbers[issueNumber] = true
				issueNumbers = append(issueNumbers, issueNumber)
			}
		}
	}

	return issueNumbers
}
