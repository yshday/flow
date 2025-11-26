/**
 * IssueTreeView
 * 이슈 트리 뷰 컴포넌트
 */

import { useState, useMemo } from 'react'
import type { Issue } from '../types'
import { IssueTreeItem } from './IssueTreeItem'

export interface IssueTreeViewProps {
  projectId: number
  issues: Issue[]
  isLoading?: boolean
  onIssueClick?: (issue: Issue) => void
  onCreateClick?: () => void
}

export function IssueTreeView({
  projectId,
  issues,
  isLoading = false,
  onIssueClick,
  onCreateClick,
}: IssueTreeViewProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterPriority, setFilterPriority] = useState<string>('')

  // Build issue tree from flat list
  const issuesWithChildren = useMemo(() => {
    return issues.map((issue) => {
      // For epics, populate epic_issues array with child issues
      if (issue.issue_type === 'epic') {
        const epicIssues = issues.filter((i) => i.epic_id === issue.id)
        return { ...issue, epic_issues: epicIssues }
      }

      // For regular issues, populate subtasks array
      const subtasks = issues.filter((i) => i.parent_issue_id === issue.id)
      return { ...issue, subtasks }
    })
  }, [issues])

  // Build issue tree: only show root-level issues (no parent_issue_id and not in an epic)
  const rootIssues = useMemo(() => {
    return issuesWithChildren.filter((issue) => !issue.parent_issue_id && !issue.epic_id)
  }, [issuesWithChildren])

  // Sort issues: pinned first, then by issue number
  const sortedRootIssues = useMemo(() => {
    return rootIssues.sort((a, b) => {
      if (a.is_pinned && !b.is_pinned) return -1
      if (!a.is_pinned && b.is_pinned) return 1
      return b.issue_number - a.issue_number // Descending order (newest first)
    })
  }, [rootIssues])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div style={{ color: 'var(--flow-text-muted, #666666)' }}>로딩 중...</div>
      </div>
    )
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-10">
        <h2
          className="text-2xl font-bold"
          style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}
        >
          이슈
        </h2>
        {onCreateClick && (
          <button
            onClick={onCreateClick}
            className="px-4 py-2 rounded-md transition-colors"
            style={{
              backgroundColor: '#2563eb',
              color: 'white',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.backgroundColor = '#1d4ed8'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.backgroundColor = '#2563eb'
            }}
          >
            새 이슈
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="mb-12 flex flex-wrap gap-4">
        <input
          type="text"
          placeholder="검색..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1 min-w-[200px] px-3 py-2 border rounded-md"
          style={{
            backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)',
            borderColor: 'var(--flow-border, #333333)',
            color: 'var(--flow-text-primary, #f0f0f0)',
          }}
        />

        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value)}
          className="px-3 py-2 border rounded-md"
          style={{
            backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)',
            borderColor: 'var(--flow-border, #333333)',
            color: 'var(--flow-text-primary, #f0f0f0)',
          }}
        >
          <option value="">모든 상태</option>
          <option value="open">열림</option>
          <option value="in_progress">진행 중</option>
          <option value="closed">닫힘</option>
        </select>

        <select
          value={filterPriority}
          onChange={(e) => setFilterPriority(e.target.value)}
          className="px-3 py-2 border rounded-md"
          style={{
            backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)',
            borderColor: 'var(--flow-border, #333333)',
            color: 'var(--flow-text-primary, #f0f0f0)',
          }}
        >
          <option value="">모든 우선순위</option>
          <option value="low">낮음</option>
          <option value="medium">보통</option>
          <option value="high">높음</option>
          <option value="urgent">긴급</option>
        </select>
      </div>

      {/* Issue Tree */}
      <div
        className="rounded-lg border"
        style={{
          backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)',
          borderColor: 'var(--flow-border, #333333)',
        }}
      >
        {sortedRootIssues.length === 0 ? (
          <div className="text-center py-12">
            <p
              className="mb-4"
              style={{ color: 'var(--flow-text-muted, #666666)' }}
            >
              이슈가 없습니다.
            </p>
            {onCreateClick && (
              <button
                onClick={onCreateClick}
                className="hover:underline"
                style={{ color: '#3b82f6' }}
              >
                첫 번째 이슈 만들기
              </button>
            )}
          </div>
        ) : (
          <div className="divide-y" style={{ borderColor: 'var(--flow-border, #333333)' }}>
            {sortedRootIssues.map((issue) => (
              <IssueTreeItem
                key={issue.id}
                issue={issue}
                projectId={projectId}
                onIssueClick={onIssueClick}
              />
            ))}
          </div>
        )}
      </div>

      {/* Issue Count */}
      <div
        className="mt-4 text-sm"
        style={{ color: 'var(--flow-text-muted, #666666)' }}
      >
        총 {issues.length}개의 이슈
      </div>
    </div>
  )
}
