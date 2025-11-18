-- Performance optimization indexes for search and statistics features

-- Issues table indexes
CREATE INDEX idx_issues_priority ON issues(priority);
CREATE INDEX idx_issues_reporter_id ON issues(reporter_id);
CREATE INDEX idx_issues_milestone_id ON issues(milestone_id);
CREATE INDEX idx_issues_updated_at ON issues(updated_at DESC);
CREATE INDEX idx_issues_created_at ON issues(created_at DESC);

-- Composite indexes for common query patterns
CREATE INDEX idx_issues_status_priority ON issues(status, priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_issues_project_status ON issues(project_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_issues_project_updated ON issues(project_id, updated_at DESC) WHERE deleted_at IS NULL;

-- Text search performance (trigram indexes for ILIKE searches)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_issues_title_trgm ON issues USING gin(title gin_trgm_ops);
CREATE INDEX idx_issues_description_trgm ON issues USING gin(description gin_trgm_ops);

-- Projects table indexes
CREATE INDEX idx_projects_updated_at ON projects(updated_at DESC);
CREATE INDEX idx_projects_name_trgm ON projects USING gin(name gin_trgm_ops);
CREATE INDEX idx_projects_description_trgm ON projects USING gin(description gin_trgm_ops);

-- Comments table index
CREATE INDEX idx_comments_created_at ON comments(created_at DESC);

-- Activities table index
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_user_created ON activities(user_id, created_at DESC);

-- Milestones table index
CREATE INDEX idx_milestones_due_date ON milestones(due_date ASC NULLS LAST);
CREATE INDEX idx_milestones_project_status ON milestones(project_id, status);
