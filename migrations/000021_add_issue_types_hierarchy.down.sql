-- Drop indexes
DROP INDEX IF EXISTS idx_issues_parent_type;
DROP INDEX IF EXISTS idx_issues_epic_type;
DROP INDEX IF EXISTS idx_issues_epic_id;
DROP INDEX IF EXISTS idx_issues_parent_issue_id;
DROP INDEX IF EXISTS idx_issues_issue_type;

-- Drop constraints
ALTER TABLE issues DROP CONSTRAINT IF EXISTS subtask_cannot_have_epic;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS epic_cannot_have_parent;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS epic_must_be_epic_type;
ALTER TABLE issues DROP CONSTRAINT IF EXISTS subtask_must_have_parent;

-- Drop columns
ALTER TABLE issues DROP COLUMN IF EXISTS epic_id;
ALTER TABLE issues DROP COLUMN IF EXISTS parent_issue_id;
ALTER TABLE issues DROP COLUMN IF EXISTS issue_type;
