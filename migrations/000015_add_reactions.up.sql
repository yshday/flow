-- Create reactions table for emoji reactions on issues and comments
CREATE TABLE IF NOT EXISTS reactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('issue', 'comment')),
    entity_id INTEGER NOT NULL,
    emoji VARCHAR(50) NOT NULL CHECK (emoji IN (
        'thumbs_up', 'thumbs_down', 'laugh', 'hooray', 'confused', 'heart', 'rocket', 'eyes'
    )),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one user can only add the same emoji once per entity
    UNIQUE(user_id, entity_type, entity_id, emoji)
);

-- Index for fast lookups by entity
CREATE INDEX idx_reactions_entity ON reactions(entity_type, entity_id);

-- Index for user reactions
CREATE INDEX idx_reactions_user ON reactions(user_id);
