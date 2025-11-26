package models

import (
	"encoding/json"
	"time"
)

// ProjectTemplateConfig 프로젝트 템플릿 설정
type ProjectTemplateConfig struct {
	Columns    []ColumnConfig `json:"columns"`
	Labels     []LabelConfig  `json:"labels"`
	Milestones []MilestoneConfig `json:"milestones,omitempty"`
}

// ColumnConfig 컬럼 설정
type ColumnConfig struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
	WIPLimit *int   `json:"wip_limit,omitempty"`
}

// LabelConfig 라벨 설정
type LabelConfig struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// MilestoneConfig 마일스톤 설정
type MilestoneConfig struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// ProjectTemplate 프로젝트 템플릿
type ProjectTemplate struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	IsSystem    bool                   `json:"is_system"`
	CreatedBy   *int                   `json:"created_by,omitempty"`
	Config      ProjectTemplateConfig  `json:"config"`
	ConfigRaw   json.RawMessage        `json:"-"` // DB에서 읽을 때 사용
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// IssueTemplate 이슈 템플릿
type IssueTemplate struct {
	ID              int       `json:"id"`
	ProjectID       int       `json:"project_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Content         string    `json:"content"`
	DefaultPriority string    `json:"default_priority"`
	DefaultLabels   []int     `json:"default_labels"`
	Position        int       `json:"position"`
	IsActive        bool      `json:"is_active"`
	CreatedBy       *int      `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateProjectTemplateRequest 프로젝트 템플릿 생성 요청
type CreateProjectTemplateRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=100"`
	Description string                `json:"description"`
	Config      ProjectTemplateConfig `json:"config" validate:"required"`
}

// CreateIssueTemplateRequest 이슈 템플릿 생성 요청
type CreateIssueTemplateRequest struct {
	Name            string `json:"name" validate:"required,min=1,max=100"`
	Description     string `json:"description"`
	Content         string `json:"content"`
	DefaultPriority string `json:"default_priority"`
	DefaultLabels   []int  `json:"default_labels"`
	Position        int    `json:"position"`
}

// UpdateIssueTemplateRequest 이슈 템플릿 수정 요청
type UpdateIssueTemplateRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Content         *string `json:"content"`
	DefaultPriority *string `json:"default_priority"`
	DefaultLabels   []int   `json:"default_labels"`
	Position        *int    `json:"position"`
	IsActive        *bool   `json:"is_active"`
}
