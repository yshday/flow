-- Tasklist items for issues (checkbox functionality)
CREATE TABLE IF NOT EXISTS tasklist_items (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    content VARCHAR(500) NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0,
    completed_at TIMESTAMP,
    completed_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tasklist_items_issue_id ON tasklist_items(issue_id);
CREATE INDEX idx_tasklist_items_position ON tasklist_items(issue_id, position);

-- Comments
COMMENT ON TABLE tasklist_items IS 'Checklist items within issues';
COMMENT ON COLUMN tasklist_items.content IS 'The text content of the task';
COMMENT ON COLUMN tasklist_items.is_completed IS 'Whether the task is checked/completed';
COMMENT ON COLUMN tasklist_items.position IS 'Order of the task within the issue';
COMMENT ON COLUMN tasklist_items.completed_at IS 'When the task was completed';
COMMENT ON COLUMN tasklist_items.completed_by IS 'Who completed the task';
