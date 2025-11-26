-- Add issue_type column
ALTER TABLE issues ADD COLUMN issue_type VARCHAR(20) NOT NULL DEFAULT 'task'
    CHECK (issue_type IN ('bug', 'improvement', 'epic', 'feature', 'task', 'subtask'));

-- Add parent_issue_id for subtasks (self-referencing)
ALTER TABLE issues ADD COLUMN parent_issue_id INTEGER REFERENCES issues(id) ON DELETE CASCADE;

-- Add epic_id for grouping issues under epics
ALTER TABLE issues ADD COLUMN epic_id INTEGER REFERENCES issues(id) ON DELETE SET NULL;

-- Add constraint: subtasks must have a parent
ALTER TABLE issues ADD CONSTRAINT subtask_must_have_parent
    CHECK (issue_type != 'subtask' OR parent_issue_id IS NOT NULL);

-- Add constraint: only epics can be epics (prevent non-epic issues from being used as epics)
ALTER TABLE issues ADD CONSTRAINT epic_must_be_epic_type
    CHECK (epic_id IS NULL OR issue_type = 'epic' OR epic_id IN (SELECT id FROM issues WHERE issue_type = 'epic'));

-- Add constraint: epics cannot have parents (they're top-level)
ALTER TABLE issues ADD CONSTRAINT epic_cannot_have_parent
    CHECK (issue_type != 'epic' OR parent_issue_id IS NULL);

-- Add constraint: subtasks cannot have epics (they belong to their parent)
ALTER TABLE issues ADD CONSTRAINT subtask_cannot_have_epic
    CHECK (issue_type != 'subtask' OR epic_id IS NULL);

-- Create indexes for performance
CREATE INDEX idx_issues_issue_type ON issues(issue_type);
CREATE INDEX idx_issues_parent_issue_id ON issues(parent_issue_id);
CREATE INDEX idx_issues_epic_id ON issues(epic_id);

-- Create composite index for common queries
CREATE INDEX idx_issues_epic_type ON issues(epic_id, issue_type) WHERE epic_id IS NOT NULL;
CREATE INDEX idx_issues_parent_type ON issues(parent_issue_id, issue_type) WHERE parent_issue_id IS NOT NULL;
