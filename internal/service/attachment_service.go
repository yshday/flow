package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/pkg/errors"
	"github.com/yourusername/issue-tracker/pkg/storage"
)

// Allowed file types for upload (whitelist)
var allowedMIMETypes = map[string]bool{
	// Documents
	"application/pdf":                                                     true,
	"application/msword":                                                  true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":                                            true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":   true,
	"application/vnd.ms-powerpoint":                                       true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":                                                          true,
	"text/csv":                                                            true,
	// Images
	"image/jpeg":                                                          true,
	"image/png":                                                           true,
	"image/gif":                                                           true,
	"image/webp":                                                          true,
	"image/svg+xml":                                                       true,
	// Archives
	"application/zip":                                                     true,
	"application/x-rar-compressed":                                        true,
	"application/x-7z-compressed":                                         true,
	"application/x-tar":                                                   true,
	"application/gzip":                                                    true,
	// Code/Text
	"text/html":                                                           true,
	"text/css":                                                            true,
	"text/javascript":                                                     true,
	"application/json":                                                    true,
	"application/xml":                                                     true,
	"text/xml":                                                            true,
	"text/markdown":                                                       true,
}

// Dangerous file extensions that should be blocked
var blockedExtensions = map[string]bool{
	".exe":  true,
	".bat":  true,
	".cmd":  true,
	".com":  true,
	".pif":  true,
	".scr":  true,
	".vbs":  true,
	".js":   true, // Executable JavaScript
	".jar":  true,
	".app":  true,
	".deb":  true,
	".rpm":  true,
	".dmg":  true,
	".pkg":  true,
	".run":  true,
	".sh":   true, // Shell scripts
	".bash": true,
	".ps1":  true, // PowerShell
}

// AttachmentService handles attachment business logic
type AttachmentService struct {
	attachmentRepo *repository.AttachmentRepository
	issueRepo      *repository.IssueRepository
	authService    *AuthorizationService
	storage        storage.Storage
	maxFileSize    int64
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(
	attachmentRepo *repository.AttachmentRepository,
	issueRepo *repository.IssueRepository,
	authService *AuthorizationService,
	storage storage.Storage,
	maxFileSize int64,
) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		issueRepo:      issueRepo,
		authService:    authService,
		storage:        storage,
		maxFileSize:    maxFileSize,
	}
}


// formatFileSize converts bytes to human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// validateFileType checks if the file type is allowed using magic number detection
func (s *AttachmentService) validateFileType(file multipart.File, filename string) (string, error) {
	// 1. Check file extension first (fast check)
	ext := strings.ToLower(filepath.Ext(filename))
	if blockedExtensions[ext] {
		return "", errors.NewValidationError(fmt.Sprintf("file type not allowed: %s files are blocked for security reasons", ext))
	}

	// 2. Detect actual MIME type using file magic numbers (prevents spoofing)
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", errors.NewInternalError("failed to read file for type detection", err)
	}

	// Reset file pointer to beginning so the file can be saved later
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", errors.NewInternalError("failed to reset file pointer", err)
	}

	// Detect content type from actual file content
	actualContentType := http.DetectContentType(buffer[:n])

	// 3. Extract base MIME type (strip charset and other parameters)
	// e.g., "text/plain; charset=utf-8" -> "text/plain"
	baseMIMEType := actualContentType
	if idx := strings.Index(actualContentType, ";"); idx != -1 {
		baseMIMEType = strings.TrimSpace(actualContentType[:idx])
	}

	// 4. Validate base MIME type against whitelist
	if !allowedMIMETypes[baseMIMEType] {
		return "", errors.NewValidationError(fmt.Sprintf("file type not allowed: detected type '%s' is not in the allowed list", actualContentType))
	}

	return actualContentType, nil
}

// Upload uploads a file attachment for an issue
func (s *AttachmentService) Upload(ctx context.Context, issueID int, userID int, file multipart.File, header *multipart.FileHeader) (*models.Attachment, error) {
	// Get issue to verify it exists and get project ID
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("issue with ID %d not found", issueID))
	}

	// Check if user has write permission (blocks viewers)
	if err := s.authService.CheckWritePermission(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	// Validate file size
	if header.Size > s.maxFileSize {
		return nil, errors.NewFileTooLargeError(fmt.Sprintf("file size %s exceeds maximum allowed size of %s",
			formatFileSize(header.Size), formatFileSize(s.maxFileSize)))
	}

	// Validate file type using magic number detection
	actualContentType, err := s.validateFileType(file, header.Filename)
	if err != nil {
		return nil, err
	}

	// Save file to storage
	_, storageKey, err := s.storage.Save(file, header.Filename)
	if err != nil {
		return nil, errors.NewInternalError("failed to save file to storage", err)
	}

	// Create attachment record
	attachment := &models.Attachment{
		IssueID:          issueID,
		UserID:           userID,
		StorageKey:       storageKey,
		OriginalFilename: header.Filename,
		FileSize:         header.Size,
		ContentType:      actualContentType,
	}

	created, err := s.attachmentRepo.Create(ctx, attachment)
	if err != nil {
		// Clean up the file if database insert fails
		cleanupErr := s.storage.Delete(storageKey)
		if cleanupErr != nil {
			log.Printf("ERROR: Failed to clean up file %s after database error: %v", storageKey, cleanupErr)
		}
		return nil, err
	}

	return created, nil
}

// Get retrieves an attachment by ID
func (s *AttachmentService) Get(ctx context.Context, id int, userID int) (*models.Attachment, error) {
	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get issue to verify access
	issue, err := s.issueRepo.GetByID(ctx, attachment.IssueID)
	if err != nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("issue with ID %d not found", attachment.IssueID))
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	return attachment, nil
}

// Download retrieves the file content for an attachment
func (s *AttachmentService) Download(ctx context.Context, id int, userID int) (io.ReadCloser, *models.Attachment, error) {
	attachment, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}

	file, err := s.storage.Get(attachment.StorageKey)
	if err != nil {
		return nil, nil, errors.NewInternalError("failed to retrieve file from storage", err)
	}

	return file, attachment, nil
}

// Delete removes an attachment
func (s *AttachmentService) Delete(ctx context.Context, id int, userID int) error {
	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get issue to verify access
	issue, err := s.issueRepo.GetByID(ctx, attachment.IssueID)
	if err != nil {
		return errors.NewNotFoundError(fmt.Sprintf("issue with ID %d not found", attachment.IssueID))
	}

	// Allow deletion if user is attachment owner
	if attachment.UserID == userID {
		err = s.attachmentRepo.Delete(ctx, id)
		if err != nil {
			return err
		}

		// Delete physical file (errors are logged but not returned to avoid inconsistent state)
		deleteErr := s.storage.Delete(attachment.StorageKey)
		if deleteErr != nil {
			log.Printf("WARNING: Failed to delete physical file %s for attachment ID %d: %v", attachment.StorageKey, id, deleteErr)
		}

		return nil
	}

	// Otherwise, check if user has admin permission (only admins/owners can delete others' attachments)
	if err := s.authService.CheckAdminPermission(ctx, issue.ProjectID, userID); err != nil {
		return err
	}

	// Soft delete in database
	err = s.attachmentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Delete physical file (errors are logged but not returned to avoid inconsistent state)
	deleteErr := s.storage.Delete(attachment.StorageKey)
	if deleteErr != nil {
		log.Printf("WARNING: Failed to delete physical file %s for attachment ID %d: %v", attachment.StorageKey, id, deleteErr)
	}

	return nil
}

// ListByIssueID retrieves all attachments for an issue
func (s *AttachmentService) ListByIssueID(ctx context.Context, issueID int, userID int) ([]*models.Attachment, error) {
	// Get issue to verify access
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("issue with ID %d not found", issueID))
	}

	// Check if user has access to the project
	if err := s.authService.CheckProjectAccess(ctx, issue.ProjectID, userID); err != nil {
		return nil, err
	}

	attachments, err := s.attachmentRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	return attachments, nil
}
