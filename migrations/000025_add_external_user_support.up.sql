-- Add external user support for SSO/OAuth integration
-- This allows users from external systems (like jmember) to be linked to Flow users

ALTER TABLE users ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS external_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255);

-- Create unique index for external_id + external_provider combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_external_id_provider
ON users (external_id, external_provider)
WHERE external_id IS NOT NULL AND external_provider IS NOT NULL;

-- Make password_hash optional for external users
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
