-- Drop pinned issues columns
DROP INDEX IF EXISTS idx_issues_pinned;
ALTER TABLE issues DROP COLUMN IF EXISTS pinned_by_user_id;
ALTER TABLE issues DROP COLUMN IF EXISTS pinned_at;
ALTER TABLE issues DROP COLUMN IF EXISTS is_pinned;
