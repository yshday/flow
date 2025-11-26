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

// TemplateHandler handles template HTTP requests
type TemplateHandler struct {
	templateService *service.TemplateService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// ==================== Project Templates ====================

// ListProjectTemplates godoc
// @Summary 프로젝트 템플릿 목록 조회
// @Description 사용 가능한 모든 프로젝트 템플릿을 조회합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ProjectTemplate
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates/projects [get]
func (h *TemplateHandler) ListProjectTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.templateService.ListProjectTemplates(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get project templates")
		return
	}

	respondJSON(w, http.StatusOK, templates)
}

// GetProjectTemplate godoc
// @Summary 프로젝트 템플릿 상세 조회
// @Description 특정 프로젝트 템플릿의 상세 정보를 조회합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "템플릿 ID"
// @Success 200 {object} models.ProjectTemplate
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/templates/projects/{id} [get]
func (h *TemplateHandler) GetProjectTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.templateService.GetProjectTemplate(r.Context(), id)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Project template not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get project template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// ==================== Issue Templates ====================

// ListIssueTemplates godoc
// @Summary 이슈 템플릿 목록 조회
// @Description 프로젝트의 모든 이슈 템플릿을 조회합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "프로젝트 ID"
// @Param active query bool false "활성 템플릿만 조회"
// @Success 200 {array} models.IssueTemplate
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{projectId}/templates/issues [get]
func (h *TemplateHandler) ListIssueTemplates(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// active 파라미터 확인
	activeOnly := r.URL.Query().Get("active") == "true"

	var templates []*models.IssueTemplate

	if activeOnly {
		templates, err = h.templateService.ListActiveIssueTemplates(r.Context(), projectID)
	} else {
		templates, err = h.templateService.ListIssueTemplates(r.Context(), projectID)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get issue templates")
		return
	}

	respondJSON(w, http.StatusOK, templates)
}

// GetIssueTemplate godoc
// @Summary 이슈 템플릿 상세 조회
// @Description 특정 이슈 템플릿의 상세 정보를 조회합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "프로젝트 ID"
// @Param templateId path int true "템플릿 ID"
// @Success 200 {object} models.IssueTemplate
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{projectId}/templates/issues/{templateId} [get]
func (h *TemplateHandler) GetIssueTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := r.PathValue("templateId")
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.templateService.GetIssueTemplate(r.Context(), templateID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue template not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get issue template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// CreateIssueTemplate godoc
// @Summary 이슈 템플릿 생성
// @Description 새로운 이슈 템플릿을 생성합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "프로젝트 ID"
// @Param body body models.CreateIssueTemplateRequest true "이슈 템플릿 정보"
// @Success 201 {object} models.IssueTemplate
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{projectId}/templates/issues [post]
func (h *TemplateHandler) CreateIssueTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(int)

	projectIDStr := r.PathValue("projectId")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.CreateIssueTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Template name is required")
		return
	}

	template, err := h.templateService.CreateIssueTemplate(r.Context(), projectID, &req, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create issue template")
		return
	}

	respondJSON(w, http.StatusCreated, template)
}

// UpdateIssueTemplate godoc
// @Summary 이슈 템플릿 수정
// @Description 기존 이슈 템플릿을 수정합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "프로젝트 ID"
// @Param templateId path int true "템플릿 ID"
// @Param body body models.UpdateIssueTemplateRequest true "수정할 정보"
// @Success 200 {object} models.IssueTemplate
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{projectId}/templates/issues/{templateId} [put]
func (h *TemplateHandler) UpdateIssueTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := r.PathValue("templateId")
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req models.UpdateIssueTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	template, err := h.templateService.UpdateIssueTemplate(r.Context(), templateID, &req)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue template not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update issue template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// DeleteIssueTemplate godoc
// @Summary 이슈 템플릿 삭제
// @Description 이슈 템플릿을 삭제합니다
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path int true "프로젝트 ID"
// @Param templateId path int true "템플릿 ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{projectId}/templates/issues/{templateId} [delete]
func (h *TemplateHandler) DeleteIssueTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := r.PathValue("templateId")
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	err = h.templateService.DeleteIssueTemplate(r.Context(), templateID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Issue template not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete issue template")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
