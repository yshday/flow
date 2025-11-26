-- Drop indexes
DROP INDEX IF EXISTS idx_issue_templates_is_active;
DROP INDEX IF EXISTS idx_issue_templates_project_id;
DROP INDEX IF EXISTS idx_project_templates_is_system;

-- Drop tables
DROP TABLE IF EXISTS issue_templates;
DROP TABLE IF EXISTS project_templates;
