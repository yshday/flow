-- Add 'in_progress' to issue status check constraint

-- Drop existing constraint
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_status_check;

-- Add new constraint with in_progress
ALTER TABLE issues ADD CONSTRAINT issues_status_check
    CHECK (status IN ('open', 'in_progress', 'closed'));
