-- Rollback performance indexes

-- Issues table indexes
DROP INDEX IF EXISTS idx_issues_priority;
DROP INDEX IF EXISTS idx_issues_reporter_id;
DROP INDEX IF EXISTS idx_issues_milestone_id;
DROP INDEX IF EXISTS idx_issues_updated_at;
DROP INDEX IF EXISTS idx_issues_created_at;

-- Composite indexes
DROP INDEX IF EXISTS idx_issues_status_priority;
DROP INDEX IF EXISTS idx_issues_project_status;
DROP INDEX IF EXISTS idx_issues_project_updated;

-- Text search indexes
DROP INDEX IF EXISTS idx_issues_title_trgm;
DROP INDEX IF EXISTS idx_issues_description_trgm;

-- Projects table indexes
DROP INDEX IF EXISTS idx_projects_updated_at;
DROP INDEX IF EXISTS idx_projects_name_trgm;
DROP INDEX IF EXISTS idx_projects_description_trgm;

-- Comments table index
DROP INDEX IF EXISTS idx_comments_created_at;

-- Activities table indexes
DROP INDEX IF EXISTS idx_activities_user_id;
DROP INDEX IF EXISTS idx_activities_user_created;

-- Milestones table indexes
DROP INDEX IF EXISTS idx_milestones_due_date;
DROP INDEX IF EXISTS idx_milestones_project_status;

-- Drop extension (only if not used elsewhere)
-- DROP EXTENSION IF EXISTS pg_trgm;
