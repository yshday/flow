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

// ProjectMemberHandler handles project member HTTP requests
type ProjectMemberHandler struct {
	memberService *service.ProjectMemberService
}

// NewProjectMemberHandler creates a new project member handler
func NewProjectMemberHandler(memberService *service.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		memberService: memberService,
	}
}

// AddMember handles adding a member to a project
func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.memberService.AddMember(r.Context(), projectID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project or user not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to add member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListMembers handles listing all members of a project
func (h *ProjectMemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	members, err := h.memberService.ListMembers(r.Context(), projectID, userID)
	if err != nil {
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list members")
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// UpdateMemberRole handles updating a member's role
func (h *ProjectMemberHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	memberUserIDStr := r.PathValue("userId")
	memberUserID, err := strconv.Atoi(memberUserIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.memberService.UpdateMemberRole(r.Context(), projectID, memberUserID, &req, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Member not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update member role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember handles removing a member from a project
func (h *ProjectMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	memberUserIDStr := r.PathValue("userId")
	memberUserID, err := strconv.Atoi(memberUserIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.memberService.RemoveMember(r.Context(), projectID, memberUserID, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Member not found")
			return
		}
		if err == pkgerrors.ErrForbidden {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == pkgerrors.ErrValidation {
			respondError(w, http.StatusBadRequest, "Cannot remove project owner")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
