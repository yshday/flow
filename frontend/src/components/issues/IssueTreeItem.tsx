import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Issue } from '../../types';
import { getStatusColor, getPriorityColor, getIssueTypeIcon } from '../../lib/utils';

interface IssueTreeItemProps {
  issue: Issue;
  level?: number;
  projectId: number;
}

export default function IssueTreeItem({ issue, level = 0, projectId }: IssueTreeItemProps) {
  const navigate = useNavigate();
  const [isExpanded, setIsExpanded] = useState(true);

  const hasChildren =
    (issue.issue_type === 'epic' && issue.epic_issues && issue.epic_issues.length > 0) ||
    (issue.subtasks && issue.subtasks.length > 0);

  const children = issue.issue_type === 'epic' ? issue.epic_issues : issue.subtasks;

  const handleClick = (e: React.MouseEvent) => {
    // Prevent navigation when clicking the expand/collapse button
    if ((e.target as HTMLElement).closest('.expand-button')) {
      return;
    }
    navigate(`/projects/${projectId}/issues/${issue.id}`);
  };

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsExpanded(!isExpanded);
  };

  return (
    <div>
      <div
        className={`flex items-center p-2 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer group border-l-2 ${
          issue.is_pinned ? 'border-l-yellow-400' : 'border-l-transparent'
        }`}
        style={{ paddingLeft: `${level * 1.5 + 0.5}rem` }}
        onClick={handleClick}
      >
        {/* Expand/Collapse Button */}
        {hasChildren && (
          <button
            className="expand-button mr-2 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
            onClick={handleToggle}
          >
            <svg
              className={`w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform ${
                isExpanded ? 'rotate-90' : ''
              }`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        )}

        {!hasChildren && <div className="w-6" />}

        {/* Issue Type Icon */}
        <span className="mr-2 text-lg">{getIssueTypeIcon(issue.issue_type)}</span>

        {/* Issue Key */}
        <span className="mr-2 text-xs font-mono text-gray-600 dark:text-gray-400 shrink-0">
          {issue.project_id}-{issue.issue_number}
        </span>

        {/* Issue Title */}
        <span className="flex-1 text-sm text-gray-900 dark:text-white truncate">{issue.title}</span>

        {/* Priority Badge */}
        <span
          className={`ml-2 px-2 py-0.5 text-xs rounded shrink-0 ${getPriorityColor(issue.priority)}`}
        >
          {issue.priority === 'low' && '낮음'}
          {issue.priority === 'medium' && '보통'}
          {issue.priority === 'high' && '높음'}
          {issue.priority === 'urgent' && '긴급'}
        </span>

        {/* Status Badge */}
        <span className={`ml-2 px-2 py-0.5 text-xs rounded shrink-0 ${getStatusColor(issue.status)}`}>
          {issue.status === 'open' && '열림'}
          {issue.status === 'in_progress' && '진행 중'}
          {issue.status === 'closed' && '닫힘'}
        </span>

        {/* Assignee */}
        {issue.assignee && (
          <span className="ml-2 text-xs text-gray-600 dark:text-gray-400 shrink-0">
            {issue.assignee.username}
          </span>
        )}

        {/* Pinned Icon */}
        {issue.is_pinned && (
          <span className="ml-2 text-yellow-500" title="고정됨">
            📌
          </span>
        )}
      </div>

      {/* Children */}
      {hasChildren && isExpanded && children && (
        <div>
          {children.map((child) => (
            <IssueTreeItem key={child.id} issue={child} level={level + 1} projectId={projectId} />
          ))}
        </div>
      )}
    </div>
  );
}
