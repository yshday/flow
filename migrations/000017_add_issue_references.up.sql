-- Create issue_references table for #issue_number references in issues and comments
CREATE TABLE IF NOT EXISTS issue_references (
    id SERIAL PRIMARY KEY,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('issue', 'comment')),
    source_id INTEGER NOT NULL,
    referenced_issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_type, source_id, referenced_issue_id)
);

-- Create indexes for performance
CREATE INDEX idx_issue_references_source ON issue_references(source_type, source_id);
CREATE INDEX idx_issue_references_referenced ON issue_references(referenced_issue_id);
CREATE INDEX idx_issue_references_created_at ON issue_references(created_at DESC);
