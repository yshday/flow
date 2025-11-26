/**
 * IssueTreeItem
 * 이슈 트리 아이템 컴포넌트 (재귀적)
 */

import { useState } from 'react'
import type { Issue } from '../types'
import { getIssueTypeIcon, getPriorityColor, getStatusColor } from '../utils'

export interface IssueTreeItemProps {
  issue: Issue
  level?: number
  projectId: number
  onIssueClick?: (issue: Issue) => void
}

export function IssueTreeItem({
  issue,
  level = 0,
  projectId,
  onIssueClick,
}: IssueTreeItemProps) {
  const [isExpanded, setIsExpanded] = useState(true)

  const hasChildren =
    (issue.issue_type === 'epic' && issue.epic_issues && issue.epic_issues.length > 0) ||
    (issue.subtasks && issue.subtasks.length > 0)

  const children = issue.issue_type === 'epic' ? issue.epic_issues : issue.subtasks

  const handleClick = (e: React.MouseEvent) => {
    // Prevent navigation when clicking the expand/collapse button
    if ((e.target as HTMLElement).closest('.expand-button')) {
      return
    }
    onIssueClick?.(issue)
  }

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIsExpanded(!isExpanded)
  }

  return (
    <div>
      <div
        className="flex items-center p-2 cursor-pointer group border-l-2"
        style={{
          paddingLeft: `${level * 1.5 + 0.5}rem`,
          borderLeftColor: issue.is_pinned
            ? '#facc15'
            : 'transparent',
          backgroundColor: 'transparent',
        }}
        onClick={handleClick}
        onMouseEnter={(e) => {
          e.currentTarget.style.backgroundColor = 'var(--flow-bg-hover, rgba(255, 255, 255, 0.05))'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = 'transparent'
        }}
      >
        {/* Expand/Collapse Button */}
        {hasChildren && (
          <button
            className="expand-button mr-2 p-0.5 rounded"
            onClick={handleToggle}
            style={{
              backgroundColor: 'transparent',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.backgroundColor = 'var(--flow-bg-tertiary, #2a2a2a)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.backgroundColor = 'transparent'
            }}
          >
            <svg
              className="w-4 h-4 transition-transform"
              style={{
                color: 'var(--flow-text-muted, #666666)',
                transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)',
              }}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        )}

        {!hasChildren && <div className="w-6" />}

        {/* Issue Type Icon */}
        <span className="mr-2 text-lg">{getIssueTypeIcon(issue.issue_type)}</span>

        {/* Issue Key */}
        <span
          className="mr-2 text-xs font-mono shrink-0"
          style={{ color: 'var(--flow-text-muted, #666666)' }}
        >
          {projectId}-{issue.issue_number}
        </span>

        {/* Issue Title */}
        <span
          className="flex-1 text-sm truncate"
          style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}
        >
          {issue.title}
        </span>

        {/* Priority Badge */}
        <span
          className="ml-2 px-2 py-0.5 text-xs rounded shrink-0"
          style={{
            backgroundColor: getPriorityBgColor(issue.priority),
            color: getPriorityTextColor(issue.priority),
          }}
        >
          {getPriorityLabel(issue.priority)}
        </span>

        {/* Status Badge */}
        <span
          className="ml-2 px-2 py-0.5 text-xs rounded shrink-0"
          style={{
            backgroundColor: getStatusBgColor(issue.status),
            color: getStatusTextColor(issue.status),
          }}
        >
          {getStatusLabel(issue.status)}
        </span>

        {/* Assignee */}
        {issue.assignee && (
          <span
            className="ml-2 text-xs shrink-0"
            style={{ color: 'var(--flow-text-muted, #666666)' }}
          >
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
          {children.map((child: Issue) => (
            <IssueTreeItem
              key={child.id}
              issue={child}
              level={level + 1}
              projectId={projectId}
              onIssueClick={onIssueClick}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// Helper functions
function getPriorityLabel(priority: string): string {
  const labels: Record<string, string> = {
    low: '낮음',
    medium: '보통',
    high: '높음',
    urgent: '긴급',
  }
  return labels[priority] || priority
}

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    open: '열림',
    in_progress: '진행 중',
    closed: '닫힘',
  }
  return labels[status] || status
}

function getPriorityBgColor(priority: string): string {
  const colors: Record<string, string> = {
    low: 'rgba(59, 130, 246, 0.2)',
    medium: 'rgba(168, 85, 247, 0.2)',
    high: 'rgba(251, 146, 60, 0.2)',
    urgent: 'rgba(239, 68, 68, 0.2)',
  }
  return colors[priority] || 'rgba(107, 114, 128, 0.2)'
}

function getPriorityTextColor(priority: string): string {
  const colors: Record<string, string> = {
    low: '#60a5fa',
    medium: '#c084fc',
    high: '#fb923c',
    urgent: '#ef4444',
  }
  return colors[priority] || '#9ca3af'
}

function getStatusBgColor(status: string): string {
  const colors: Record<string, string> = {
    open: 'rgba(34, 197, 94, 0.2)',
    in_progress: 'rgba(234, 179, 8, 0.2)',
    closed: 'rgba(107, 114, 128, 0.2)',
  }
  return colors[status] || 'rgba(107, 114, 128, 0.2)'
}

function getStatusTextColor(status: string): string {
  const colors: Record<string, string> = {
    open: '#22c55e',
    in_progress: '#eab308',
    closed: '#6b7280',
  }
  return colors[status] || '#9ca3af'
}
