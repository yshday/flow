-- Webhooks table: stores webhook configurations per project
CREATE TABLE webhooks (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255),  -- For HMAC signature verification
    events TEXT[] NOT NULL DEFAULT '{}',  -- Array of event types to subscribe to
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Webhook deliveries: logs of webhook delivery attempts
CREATE TABLE webhook_deliveries (
    id SERIAL PRIMARY KEY,
    webhook_id INTEGER NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status INTEGER,
    response_body TEXT,
    error_message TEXT,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_webhooks_project_id ON webhooks(project_id);
CREATE INDEX idx_webhooks_is_active ON webhooks(is_active);
CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_created_at ON webhook_deliveries(created_at);

-- Comments
COMMENT ON TABLE webhooks IS 'Webhook configurations for projects';
COMMENT ON TABLE webhook_deliveries IS 'Webhook delivery logs and attempts';
COMMENT ON COLUMN webhooks.events IS 'Array of event types: issue.created, issue.updated, issue.deleted, comment.created, etc.';
COMMENT ON COLUMN webhooks.secret IS 'Secret key for HMAC-SHA256 signature verification';
