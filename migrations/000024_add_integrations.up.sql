-- Integrations table: stores messenger integration configurations per project
CREATE TABLE integrations (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,  -- 'slack', 'discord', 'teams', etc.
    webhook_url TEXT NOT NULL,  -- The incoming webhook URL for the messenger
    channel VARCHAR(255),       -- Channel name for display purposes
    events TEXT[] NOT NULL DEFAULT '{}',  -- Array of event types to subscribe to
    is_active BOOLEAN NOT NULL DEFAULT true,
    settings JSONB DEFAULT '{}',  -- Additional settings (e.g., username, icon)
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Integration message logs: logs of message delivery attempts
CREATE TABLE integration_messages (
    id SERIAL PRIMARY KEY,
    integration_id INTEGER NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    response_status INTEGER,
    error_message TEXT,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_integrations_project_id ON integrations(project_id);
CREATE INDEX idx_integrations_type ON integrations(type);
CREATE INDEX idx_integrations_is_active ON integrations(is_active);
CREATE INDEX idx_integration_messages_integration_id ON integration_messages(integration_id);
CREATE INDEX idx_integration_messages_created_at ON integration_messages(created_at);

-- Comments
COMMENT ON TABLE integrations IS 'Messenger integration configurations for projects (Slack, Discord, etc.)';
COMMENT ON TABLE integration_messages IS 'Integration message delivery logs';
COMMENT ON COLUMN integrations.type IS 'Integration type: slack, discord, teams, custom';
COMMENT ON COLUMN integrations.webhook_url IS 'Incoming webhook URL for the messenger service';
COMMENT ON COLUMN integrations.settings IS 'JSON settings like username, icon_url, color preferences';
