package service

import (
	"context"

	"github.com/yourusername/issue-tracker/internal/models"
	"github.com/yourusername/issue-tracker/internal/repository"
)

// TemplateService handles template business logic
type TemplateService struct {
	templateRepo *repository.TemplateRepository
}

// NewTemplateService creates a new template service
func NewTemplateService(templateRepo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
	}
}

// ==================== Project Templates ====================

// ListProjectTemplates returns all available project templates
func (s *TemplateService) ListProjectTemplates(ctx context.Context) ([]*models.ProjectTemplate, error) {
	return s.templateRepo.ListProjectTemplates(ctx)
}

// GetProjectTemplate returns a project template by ID
func (s *TemplateService) GetProjectTemplate(ctx context.Context, id int) (*models.ProjectTemplate, error) {
	return s.templateRepo.GetProjectTemplate(ctx, id)
}

// CreateProjectTemplate creates a new custom project template
func (s *TemplateService) CreateProjectTemplate(ctx context.Context, req *models.CreateProjectTemplateRequest, userID int) (*models.ProjectTemplate, error) {
	template := &models.ProjectTemplate{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		IsSystem:    false, // 사용자 생성 템플릿은 항상 false
		CreatedBy:   &userID,
	}

	return s.templateRepo.CreateProjectTemplate(ctx, template)
}

// ==================== Issue Templates ====================

// ListIssueTemplates returns all issue templates for a project
func (s *TemplateService) ListIssueTemplates(ctx context.Context, projectID int) ([]*models.IssueTemplate, error) {
	return s.templateRepo.ListIssueTemplates(ctx, projectID)
}

// ListActiveIssueTemplates returns only active issue templates for a project
func (s *TemplateService) ListActiveIssueTemplates(ctx context.Context, projectID int) ([]*models.IssueTemplate, error) {
	return s.templateRepo.ListActiveIssueTemplates(ctx, projectID)
}

// GetIssueTemplate returns an issue template by ID
func (s *TemplateService) GetIssueTemplate(ctx context.Context, id int) (*models.IssueTemplate, error) {
	return s.templateRepo.GetIssueTemplate(ctx, id)
}

// CreateIssueTemplate creates a new issue template
func (s *TemplateService) CreateIssueTemplate(ctx context.Context, projectID int, req *models.CreateIssueTemplateRequest, userID int) (*models.IssueTemplate, error) {
	// 기본값 설정
	defaultPriority := req.DefaultPriority
	if defaultPriority == "" {
		defaultPriority = "medium"
	}

	defaultLabels := req.DefaultLabels
	if defaultLabels == nil {
		defaultLabels = []int{}
	}

	template := &models.IssueTemplate{
		ProjectID:       projectID,
		Name:            req.Name,
		Description:     req.Description,
		Content:         req.Content,
		DefaultPriority: defaultPriority,
		DefaultLabels:   defaultLabels,
		Position:        req.Position,
		IsActive:        true,
		CreatedBy:       &userID,
	}

	return s.templateRepo.CreateIssueTemplate(ctx, template)
}

// UpdateIssueTemplate updates an issue template
func (s *TemplateService) UpdateIssueTemplate(ctx context.Context, id int, req *models.UpdateIssueTemplateRequest) (*models.IssueTemplate, error) {
	// 기존 템플릿 조회
	template, err := s.templateRepo.GetIssueTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 변경사항 적용
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.Content != nil {
		template.Content = *req.Content
	}
	if req.DefaultPriority != nil {
		template.DefaultPriority = *req.DefaultPriority
	}
	if req.DefaultLabels != nil {
		template.DefaultLabels = req.DefaultLabels
	}
	if req.Position != nil {
		template.Position = *req.Position
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	// 업데이트 실행
	if err := s.templateRepo.UpdateIssueTemplate(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// DeleteIssueTemplate deletes an issue template
func (s *TemplateService) DeleteIssueTemplate(ctx context.Context, id int) error {
	return s.templateRepo.DeleteIssueTemplate(ctx, id)
}
