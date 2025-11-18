-- Create mentions table for @username mentions in issues and comments
CREATE TABLE IF NOT EXISTS mentions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mentioned_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('issue', 'comment')),
    entity_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, entity_type, entity_id)
);

-- Create indexes for performance
CREATE INDEX idx_mentions_user ON mentions(user_id);
CREATE INDEX idx_mentions_entity ON mentions(entity_type, entity_id);
CREATE INDEX idx_mentions_mentioned_by ON mentions(mentioned_by_user_id);
CREATE INDEX idx_mentions_created_at ON mentions(created_at DESC);
