package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/service"
	pkgerrors "github.com/yourusername/issue-tracker/pkg/errors"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// List handles listing notifications for the authenticated user
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	unreadOnly := r.URL.Query().Get("unread") == "true"

	limit := 20 // default
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // max limit
			}
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var notifications []*models.Notification
	var err error

	if unreadOnly {
		notifications, err = h.notificationService.ListUnreadByUserID(r.Context(), userID, limit, offset)
	} else {
		notifications, err = h.notificationService.ListByUserID(r.Context(), userID, limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list notifications")
		return
	}

	respondJSON(w, http.StatusOK, notifications)
}

// GetByID handles retrieving a notification by ID
func (h *NotificationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	notification, err := h.notificationService.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	respondJSON(w, http.StatusOK, notification)
}

// MarkAsRead handles marking notifications as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	var req models.MarkAsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.NotificationIDs) == 0 {
		respondError(w, http.StatusBadRequest, "notification_ids is required")
		return
	}

	err := h.notificationService.MarkAsRead(r.Context(), &req, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to mark notifications as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkAllAsRead handles marking all notifications as read for a user
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	err := h.notificationService.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles deleting a notification
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	err = h.notificationService.Delete(r.Context(), id, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Notification not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUnreadCount handles getting the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	count, err := h.notificationService.CountUnread(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count unread notifications")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}
