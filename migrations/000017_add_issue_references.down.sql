-- Drop issue_references table
DROP INDEX IF EXISTS idx_issue_references_created_at;
DROP INDEX IF EXISTS idx_issue_references_referenced;
DROP INDEX IF EXISTS idx_issue_references_source;
DROP TABLE IF EXISTS issue_references;
