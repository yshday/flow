-- Refactor attachments table
-- Combine filename and file_path into storage_key
-- Add updated_at column

-- Add new columns
ALTER TABLE attachments ADD COLUMN IF NOT EXISTS storage_key VARCHAR(500);
ALTER TABLE attachments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Migrate data: use file_path as storage_key (they are the same in current implementation)
UPDATE attachments SET storage_key = file_path WHERE storage_key IS NULL;
UPDATE attachments SET updated_at = created_at WHERE updated_at IS NULL;

-- Make storage_key NOT NULL after data migration
ALTER TABLE attachments ALTER COLUMN storage_key SET NOT NULL;
ALTER TABLE attachments ALTER COLUMN updated_at SET NOT NULL;

-- Make old columns nullable for backward compatibility
-- This allows new records to be created without these fields
ALTER TABLE attachments ALTER COLUMN filename DROP NOT NULL;
ALTER TABLE attachments ALTER COLUMN file_path DROP NOT NULL;

-- Drop old columns (keeping for now for backward compatibility)
-- We'll drop these in a future migration after ensuring everything works
-- ALTER TABLE attachments DROP COLUMN filename;
-- ALTER TABLE attachments DROP COLUMN file_path;

-- Update comments
COMMENT ON COLUMN attachments.storage_key IS 'Internal storage key (not exposed in API)';
COMMENT ON COLUMN attachments.updated_at IS 'Last updated timestamp';
