DROP TRIGGER IF EXISTS trigger_set_issue_number ON issues;
DROP FUNCTION IF EXISTS set_issue_number();
DROP FUNCTION IF EXISTS get_next_issue_number(INTEGER);
DROP TABLE IF EXISTS project_issue_counters CASCADE;
