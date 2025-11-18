-- Test query performance with EXPLAIN ANALYZE
-- Run with: docker exec -i <container> psql -U postgres -d issue_tracker < test_query_performance.sql

-- Test 1: Issue search with text query (should use GIN trigram index)
EXPLAIN ANALYZE
SELECT
    i.id, i.project_id, p.key as project_key, i.issue_number,
    i.title, COALESCE(i.description, '') as description,
    i.status, i.priority, i.assignee_id, i.reporter_id,
    i.created_at, i.updated_at
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.deleted_at IS NULL
  AND (i.title ILIKE '%bug%' OR i.description ILIKE '%bug%')
ORDER BY i.updated_at DESC
LIMIT 20 OFFSET 0;

-- Test 2: Issue search with status and priority filters (should use composite index)
EXPLAIN ANALYZE
SELECT
    i.id, i.project_id, p.key as project_key, i.issue_number,
    i.title, COALESCE(i.description, '') as description,
    i.status, i.priority, i.assignee_id, i.reporter_id,
    i.created_at, i.updated_at
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.deleted_at IS NULL
  AND i.status = 'open'
  AND i.priority = 'high'
ORDER BY i.updated_at DESC
LIMIT 20 OFFSET 0;

-- Test 3: Project statistics query (should use various indexes)
EXPLAIN ANALYZE
SELECT
    COUNT(CASE WHEN i.deleted_at IS NULL THEN 1 END) as total_issues,
    COUNT(CASE WHEN i.status = 'open' AND i.deleted_at IS NULL THEN 1 END) as open_issues,
    COUNT(CASE WHEN i.status = 'closed' AND i.deleted_at IS NULL THEN 1 END) as closed_issues,
    COUNT(CASE WHEN i.priority = 'critical' AND i.deleted_at IS NULL THEN 1 END) as critical_issues,
    COUNT(CASE WHEN i.priority = 'high' AND i.deleted_at IS NULL THEN 1 END) as high_issues,
    COUNT(CASE WHEN i.priority = 'medium' AND i.deleted_at IS NULL THEN 1 END) as medium_issues,
    COUNT(CASE WHEN i.priority = 'low' AND i.deleted_at IS NULL THEN 1 END) as low_issues
FROM issues i
WHERE i.project_id = 1;

-- Test 4: Issue statistics with time-based filtering (should use created_at/updated_at indexes)
EXPLAIN ANALYZE
SELECT
    COUNT(*) as issues_created_last_30_days
FROM issues
WHERE deleted_at IS NULL
  AND project_id = 1
  AND created_at >= CURRENT_TIMESTAMP - INTERVAL '30 days';

-- Test 5: Project search with text query (should use GIN trigram index)
EXPLAIN ANALYZE
SELECT
    id, name, key,
    COALESCE(description, '') as description,
    owner_id, created_at
FROM projects
WHERE (name ILIKE '%test%' OR description ILIKE '%test%' OR key ILIKE '%test%')
ORDER BY updated_at DESC
LIMIT 20 OFFSET 0;

-- Test 6: Issues by project with filtering (should use composite index)
EXPLAIN ANALYZE
SELECT
    i.id, i.project_id, p.key as project_key, i.issue_number,
    i.title, i.status, i.priority
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.project_id = 1
  AND i.status = 'open'
  AND i.deleted_at IS NULL
ORDER BY i.updated_at DESC;

-- Test 7: User activity statistics (should use reporter_id index)
EXPLAIN ANALYZE
SELECT COUNT(*) as issues_created
FROM issues
WHERE reporter_id = 20
  AND deleted_at IS NULL;

-- Test 8: Milestone-based filtering (should use milestone_id index)
EXPLAIN ANALYZE
SELECT
    i.id, i.title, i.status, i.priority
FROM issues i
WHERE i.milestone_id = 1
  AND i.deleted_at IS NULL
ORDER BY i.priority DESC, i.updated_at DESC;
