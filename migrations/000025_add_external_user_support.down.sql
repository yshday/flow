-- Revert external user support

DROP INDEX IF EXISTS idx_users_external_id_provider;

ALTER TABLE users DROP COLUMN IF EXISTS external_id;
ALTER TABLE users DROP COLUMN IF EXISTS external_provider;
ALTER TABLE users DROP COLUMN IF EXISTS name;

-- Restore password_hash NOT NULL (may fail if external users exist)
-- ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
