-- Create issue_watchers table for users watching/subscribing to issues
CREATE TABLE IF NOT EXISTS issue_watchers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, issue_id)
);

-- Create indexes for performance
CREATE INDEX idx_issue_watchers_user ON issue_watchers(user_id);
CREATE INDEX idx_issue_watchers_issue ON issue_watchers(issue_id);
CREATE INDEX idx_issue_watchers_created_at ON issue_watchers(created_at DESC);
