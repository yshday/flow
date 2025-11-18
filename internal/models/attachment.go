package models

import "time"

// Attachment represents a file attachment for an issue
type Attachment struct {
	ID               int        `json:"id"`
	IssueID          int        `json:"issue_id"`
	UserID           int        `json:"user_id"`
	StorageKey       string     `json:"-"` // Internal storage key (not exposed in API)
	OriginalFilename string     `json:"original_filename"`
	FileSize         int64      `json:"file_size"`
	ContentType      string     `json:"content_type"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

// AttachmentUploadRequest represents a file upload request
type AttachmentUploadRequest struct {
	IssueID int `json:"issue_id"`
}

// AttachmentResponse represents an attachment in API responses
type AttachmentResponse struct {
	ID               int       `json:"id"`
	IssueID          int       `json:"issue_id"`
	UserID           int       `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	FileSize         int64     `json:"file_size"`
	ContentType      string    `json:"content_type"`
	DownloadURL      string    `json:"download_url"`
	CreatedAt        time.Time `json:"created_at"`
}
