-- Drop issue_watchers table
DROP INDEX IF EXISTS idx_issue_watchers_created_at;
DROP INDEX IF EXISTS idx_issue_watchers_issue;
DROP INDEX IF EXISTS idx_issue_watchers_user;
DROP TABLE IF EXISTS issue_watchers;
