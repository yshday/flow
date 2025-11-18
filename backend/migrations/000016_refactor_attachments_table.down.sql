-- Rollback attachments table refactoring

-- Drop new columns
ALTER TABLE attachments DROP COLUMN IF EXISTS storage_key;
ALTER TABLE attachments DROP COLUMN IF EXISTS updated_at;
