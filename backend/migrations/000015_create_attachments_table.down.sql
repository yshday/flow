-- Drop indexes
DROP INDEX IF EXISTS idx_attachments_created_at;
DROP INDEX IF EXISTS idx_attachments_user_id;
DROP INDEX IF EXISTS idx_attachments_issue_id;

-- Drop table
DROP TABLE IF EXISTS attachments;
