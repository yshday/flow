package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/service"
	"github.com/yourusername/issue-tracker/pkg/errors"
)

// AttachmentHandler handles attachment-related HTTP requests
type AttachmentHandler struct {
	attachmentService *service.AttachmentService
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(attachmentService *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
	}
}

// getHTTPStatusCode extracts the HTTP status code from an error
func getHTTPStatusCode(err error) int {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// sanitizeFilename removes potentially dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Get base name and extension
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	// Remove path separators and null bytes
	base = strings.ReplaceAll(base, "/", "_")
	base = strings.ReplaceAll(base, "\\", "_")
	base = strings.ReplaceAll(base, "\x00", "")

	// Remove control characters and other dangerous chars
	reg := regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)
	base = reg.ReplaceAllString(base, "_")

	// Trim dots and spaces from start/end
	base = strings.Trim(base, ". ")

	// Limit length
	const maxLen = 200
	if len(base) > maxLen {
		base = base[:maxLen]
	}

	// If base is empty after sanitization, use a default name
	if base == "" {
		base = "attachment"
	}

	return base + ext
}

// Upload handles file upload
// POST /api/v1/issues/{id}/attachments
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Get issue ID from URL
	issueIDStr := r.PathValue("id")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	// Parse multipart form (32MB max memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Upload file
	attachment, err := h.attachmentService.Upload(r.Context(), issueID, userID, file, header)
	if err != nil {
		statusCode := getHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	// Convert to response format
	response := &models.AttachmentResponse{
		ID:               attachment.ID,
		IssueID:          attachment.IssueID,
		UserID:           attachment.UserID,
		OriginalFilename: attachment.OriginalFilename,
		FileSize:         attachment.FileSize,
		ContentType:      attachment.ContentType,
		DownloadURL:      fmt.Sprintf("/api/v1/attachments/%d/download", attachment.ID),
		CreatedAt:        attachment.CreatedAt,
	}

	respondJSON(w, http.StatusCreated, response)
}

// Download handles file download
// GET /api/v1/attachments/{id}/download
func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Get attachment ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid attachment ID")
		return
	}

	// Download file
	file, attachment, err := h.attachmentService.Download(r.Context(), id, userID)
	if err != nil {
		statusCode := getHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}
	defer file.Close()

	// Sanitize filename to prevent header injection attacks
	safeFilename := sanitizeFilename(attachment.OriginalFilename)

	// Set headers for download
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeFilename))
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(attachment.FileSize, 10))

	// Stream file to response
	_, err = io.Copy(w, file)
	if err != nil {
		// Can't send JSON error here as headers are already sent
		// Error will be logged by service layer
		return
	}
}

// Delete handles attachment deletion
// DELETE /api/v1/attachments/{id}
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Get attachment ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid attachment ID")
		return
	}

	// Delete attachment
	err = h.attachmentService.Delete(r.Context(), id, userID)
	if err != nil {
		statusCode := getHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Attachment deleted successfully"})
}

// List handles listing attachments for an issue
// GET /api/v1/issues/{id}/attachments
func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Get issue ID from URL
	issueIDStr := r.PathValue("id")
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid issue ID")
		return
	}

	// Get attachments
	attachments, err := h.attachmentService.ListByIssueID(r.Context(), issueID, userID)
	if err != nil {
		statusCode := getHTTPStatusCode(err)
		respondError(w, statusCode, err.Error())
		return
	}

	// Convert to response format
	responses := make([]*models.AttachmentResponse, len(attachments))
	for i, attachment := range attachments {
		responses[i] = &models.AttachmentResponse{
			ID:               attachment.ID,
			IssueID:          attachment.IssueID,
			UserID:           attachment.UserID,
			OriginalFilename: attachment.OriginalFilename,
			FileSize:         attachment.FileSize,
			ContentType:      attachment.ContentType,
			DownloadURL:      fmt.Sprintf("/api/v1/attachments/%d/download", attachment.ID),
			CreatedAt:        attachment.CreatedAt,
		}
	}

	respondJSON(w, http.StatusOK, responses)
}
