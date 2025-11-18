-- Revert to original two-state status constraint

-- Drop the three-state constraint
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_status_check;

-- Add back the original two-state constraint
ALTER TABLE issues ADD CONSTRAINT issues_status_check
    CHECK (status IN ('open', 'closed'));
