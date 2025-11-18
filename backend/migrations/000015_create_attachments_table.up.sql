-- Create attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_attachments_issue_id ON attachments(issue_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_attachments_user_id ON attachments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_attachments_created_at ON attachments(created_at DESC) WHERE deleted_at IS NULL;

-- Add comment
COMMENT ON TABLE attachments IS 'File attachments for issues';
COMMENT ON COLUMN attachments.filename IS 'Unique filename on disk (UUID-based)';
COMMENT ON COLUMN attachments.original_filename IS 'Original filename uploaded by user';
COMMENT ON COLUMN attachments.file_path IS 'Relative path to file from storage root';
COMMENT ON COLUMN attachments.file_size IS 'File size in bytes';
