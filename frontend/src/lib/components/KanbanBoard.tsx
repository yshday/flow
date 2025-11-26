/**
 * KanbanBoard
 * 칸반 보드 컴포넌트
 */

import { useCallback, useState } from 'react'
import { useFlowProject, useFlowBoardColumns, useFlowIssues, useFlowMoveIssue } from '../hooks'
import { flowIssuesApi } from '../api'
import type { Issue } from '../types'
import { IssueCompleteModal } from './IssueCompleteModal'

interface PendingComplete {
  issue: Issue
  columnId: number
}

export interface KanbanBoardProps {
  projectId: number
  onBack?: () => void
  onIssueClick?: (issue: Issue) => void
  onCreateClick?: () => void
  onSettingsClick?: () => void
}

export function KanbanBoard({
  projectId,
  onBack,
  onIssueClick,
  onCreateClick,
  onSettingsClick,
}: KanbanBoardProps) {
  const { data: project, isLoading: projectLoading } = useFlowProject(projectId)
  const { data: columns, isLoading: columnsLoading } = useFlowBoardColumns(projectId)
  const { data: issues, isLoading: issuesLoading } = useFlowIssues(projectId)
  const { mutateAsync: moveIssue } = useFlowMoveIssue()

  // 완료 모달 관련 상태
  const [pendingComplete, setPendingComplete] = useState<PendingComplete | null>(null)
  const [isCompleting, setIsCompleting] = useState(false)

  const handleDrop = useCallback(
    async (issueId: number, columnId: number, version: number) => {
      const targetColumn = columns?.find((col) => col.id === columnId)
      const columnName = targetColumn?.name.toLowerCase()
      const issue = issues?.find((i) => i.id === issueId)

      // Done 컬럼으로 이동하는 경우 모달 표시
      if (columnName === 'done' && issue) {
        setPendingComplete({ issue, columnId })
        return
      }

      let status: 'open' | 'in_progress' | 'closed' = 'open'
      if (columnName === 'in progress') {
        status = 'in_progress'
      }

      try {
        await moveIssue({
          id: issueId,
          data: { column_id: columnId, version, status },
        })
      } catch (error) {
        console.error('Failed to move issue:', error)
      }
    },
    [columns, issues, moveIssue]
  )

  // 완료 모달 확인 핸들러
  const handleCompleteConfirm = useCallback(
    async (comment: string) => {
      if (!pendingComplete) return

      setIsCompleting(true)
      try {
        // 1. 이슈를 완료 상태로 이동
        await moveIssue({
          id: pendingComplete.issue.id,
          data: {
            column_id: pendingComplete.columnId,
            version: pendingComplete.issue.version,
            status: 'closed',
          },
        })

        // 2. 댓글이 있으면 등록
        if (comment) {
          const completionComment = `✅ **이슈 완료**\n\n${comment}`
          await flowIssuesApi.createComment(pendingComplete.issue.id, completionComment)
        }

        setPendingComplete(null)
      } catch (error) {
        console.error('Failed to complete issue:', error)
      } finally {
        setIsCompleting(false)
      }
    },
    [pendingComplete, moveIssue]
  )

  // 완료 모달 취소 핸들러
  const handleCompleteCancel = useCallback(() => {
    setPendingComplete(null)
  }, [])

  const isLoading = projectLoading || columnsLoading || issuesLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">로딩 중...</div>
      </div>
    )
  }

  if (!project || !columns) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-red-500">데이터를 불러올 수 없습니다.</div>
      </div>
    )
  }

  // Group issues by column
  const issuesByColumn = new Map<number, Issue[]>()
  columns.forEach((col) => issuesByColumn.set(col.id, []))
  issues?.forEach((issue) => {
    if (issue.column_id) {
      const columnIssues = issuesByColumn.get(issue.column_id) || []
      columnIssues.push(issue)
      issuesByColumn.set(issue.column_id, columnIssues)
    }
  })

  return (
    <div className="flow-kanban-board p-4">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          {onBack && (
            <button
              onClick={onBack}
              className="text-sm hover:opacity-80"
              style={{ color: 'var(--flow-text-muted, #666666)' }}
            >
              &larr; 뒤로
            </button>
          )}
          <div>
            <h1 className="text-xl font-bold" style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}>{project.name}</h1>
            <p className="text-sm" style={{ color: 'var(--flow-text-muted, #666666)' }}>{project.key}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {onSettingsClick && (
            <button
              onClick={onSettingsClick}
              className="px-3 py-2 border rounded-md text-sm hover:opacity-80"
              style={{ color: 'var(--flow-text-secondary, #a0a0a0)', borderColor: 'var(--flow-border, #333333)' }}
              title="프로젝트 설정"
            >
              설정
            </button>
          )}
          {onCreateClick && (
            <button
              onClick={onCreateClick}
              className="px-4 py-2 rounded-md text-sm"
              style={{ backgroundColor: '#2563eb', color: 'white' }}
            >
              새 이슈
            </button>
          )}
        </div>
      </div>

      {/* Board */}
      <div className="flex gap-4 pb-4">
        {columns
          .sort((a, b) => a.position - b.position)
          .map((column) => (
            <div
              key={column.id}
              className="flex-1 min-w-0 bg-gray-100 rounded-lg p-3"
              style={{ backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)' }}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => {
                e.preventDefault()
                const data = e.dataTransfer.getData('application/json')
                if (data) {
                  const { issueId, version } = JSON.parse(data)
                  handleDrop(issueId, column.id, version)
                }
              }}
            >
              <div className="flex items-center justify-between gap-3 mb-4 pb-2 border-b" style={{ borderColor: 'var(--flow-border, #333333)' }}>
                <h3 className="font-medium" style={{ color: 'var(--flow-text-secondary, #a0a0a0)' }}>{translateColumnName(column.name)}</h3>
                <span className="text-xs px-2 py-1 rounded" style={{ backgroundColor: 'var(--flow-bg-tertiary, #252525)', color: 'var(--flow-text-muted, #666666)' }}>
                  {issuesByColumn.get(column.id)?.length || 0}
                </span>
              </div>

              <div className="space-y-2">
                {issuesByColumn.get(column.id)?.map((issue) => (
                  <div
                    key={issue.id}
                    draggable
                    onDragStart={(e) => {
                      e.dataTransfer.setData(
                        'application/json',
                        JSON.stringify({ issueId: issue.id, version: issue.version })
                      )
                    }}
                    onClick={() => onIssueClick?.(issue)}
                    className="p-3 rounded border cursor-pointer hover:border-blue-400 hover:shadow-sm transition-all"
                    style={{ backgroundColor: 'var(--flow-bg-tertiary, #252525)', borderColor: 'var(--flow-border, #333333)' }}
                  >
                    <div className="text-xs font-medium mb-1" style={{ color: '#60a5fa' }}>
                      {project.key}-{issue.issue_number}
                    </div>
                    <div className="text-sm font-medium line-clamp-2" style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}>
                      {issue.title}
                    </div>
                    <div className="flex items-center gap-2 mt-2">
                      <span
                        className={`text-xs px-2 py-0.5 rounded ${getPriorityColor(
                          issue.priority
                        )}`}
                      >
                        {translatePriority(issue.priority)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
      </div>

      {/* Issue Complete Modal */}
      {project && pendingComplete && (
        <IssueCompleteModal
          isOpen={!!pendingComplete}
          issueKey={`${project.key}-${pendingComplete.issue.issue_number}`}
          issueTitle={pendingComplete.issue.title}
          onConfirm={handleCompleteConfirm}
          onCancel={handleCompleteCancel}
          isPending={isCompleting}
        />
      )}
    </div>
  )
}

function getPriorityColor(priority: string): string {
  const colors: Record<string, string> = {
    low: 'bg-gray-100 text-gray-600',
    medium: 'bg-blue-100 text-blue-600',
    high: 'bg-orange-100 text-orange-600',
    urgent: 'bg-red-100 text-red-600',
  }
  return colors[priority] || colors.medium
}

/** 컬럼 이름 한국어 번역 */
function translateColumnName(name: string): string {
  const translations: Record<string, string> = {
    'Backlog': '백로그',
    'In Progress': '진행 중',
    'Done': '완료',
  }
  return translations[name] || name
}

/** 우선순위 한국어 번역 */
function translatePriority(priority: string): string {
  const translations: Record<string, string> = {
    'low': '낮음',
    'medium': '보통',
    'high': '높음',
    'urgent': '긴급',
  }
  return translations[priority] || priority
}
