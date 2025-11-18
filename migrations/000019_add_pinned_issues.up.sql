-- Add is_pinned column to issues table
ALTER TABLE issues ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE issues ADD COLUMN IF NOT EXISTS pinned_at TIMESTAMP;
ALTER TABLE issues ADD COLUMN IF NOT EXISTS pinned_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- Create index for pinned issues queries
CREATE INDEX idx_issues_pinned ON issues(project_id, is_pinned, pinned_at DESC) WHERE is_pinned = TRUE;
