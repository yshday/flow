-- Drop mentions table
DROP INDEX IF EXISTS idx_mentions_created_at;
DROP INDEX IF EXISTS idx_mentions_mentioned_by;
DROP INDEX IF EXISTS idx_mentions_entity;
DROP INDEX IF EXISTS idx_mentions_user;
DROP TABLE IF EXISTS mentions;
