#!/bin/bash

cd /Users/ysh/dev/flow/migrations

for file in 000001_create_users_table.up.sql \
            000002_create_refresh_tokens_table.up.sql \
            000003_create_projects_table.up.sql \
            000004_create_project_members_table.up.sql \
            000005_create_board_columns_table.up.sql \
            000006_create_milestones_table.up.sql \
            000007_create_labels_table.up.sql \
            000008_create_issues_table.up.sql \
            000009_create_issue_labels_table.up.sql \
            000010_create_comments_table.up.sql \
            000011_create_activities_table.up.sql \
            000012_create_issue_counter.up.sql \
            000013_create_notifications_table.up.sql \
            000014_add_performance_indexes.up.sql \
            000015_add_reactions.up.sql \
            000016_add_mentions.up.sql \
            000017_add_issue_references.up.sql \
            000018_add_issue_watchers.up.sql \
            000019_add_pinned_issues.up.sql
do
    echo "Running migration: $file"
    cat "$file" | docker exec -i issue-tracker-db psql -U postgres -d issuetracker
done

echo "Migrations completed!"
