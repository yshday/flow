/**
 * IssueDetail
 * 이슈 상세 뷰 컴포넌트
 */

import { useState, useCallback } from 'react'
import { useFlowIssue, useFlowUpdateIssue, useFlowComments, useFlowCreateComment, useFlowMilestones } from '../hooks'
import type { Issue, IssueStatus, IssuePriority } from '../types'
import { IssueCompleteModal } from './IssueCompleteModal'

export interface IssueDetailProps {
  issueId: number
  onBack?: () => void
  onUpdate?: (issue: Issue) => void
}

export function IssueDetail({ issueId, onBack, onUpdate }: IssueDetailProps) {
  const { data: issue, isLoading, error } = useFlowIssue(issueId)
  const { data: comments } = useFlowComments(issueId)
  const { mutateAsync: updateIssue, isPending: isUpdating } = useFlowUpdateIssue(issueId)
  const { mutateAsync: createComment, isPending: isCommenting } = useFlowCreateComment(issueId)
  const { data: milestones } = useFlowMilestones(issue?.project_id ?? 0)

  const [newComment, setNewComment] = useState('')
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')

  // 완료 모달 관련 상태
  const [showCompleteModal, setShowCompleteModal] = useState(false)
  const [isClosing, setIsClosing] = useState(false)

  // 완료 모달 확인 핸들러 (hooks는 early return 전에 선언해야 함)
  const handleCompleteConfirm = useCallback(
    async (comment: string) => {
      if (!issue) return

      setIsClosing(true)
      try {
        // 1. 이슈를 완료 상태로 변경
        const updated = await updateIssue({ status: 'closed', version: issue.version })
        onUpdate?.(updated)

        // 2. 댓글이 있으면 등록
        if (comment) {
          const completionComment = `✅ **이슈 완료**\n\n${comment}`
          await createComment(completionComment)
        }

        setShowCompleteModal(false)
      } catch (error) {
        console.error('Failed to complete issue:', error)
      } finally {
        setIsClosing(false)
      }
    },
    [issue, updateIssue, createComment, onUpdate]
  )

  // 완료 모달 취소 핸들러
  const handleCompleteCancel = useCallback(() => {
    setShowCompleteModal(false)
  }, [])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500 dark:text-gray-400">로딩 중...</div>
      </div>
    )
  }

  if (error || !issue) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-red-500 dark:text-red-400">이슈를 불러올 수 없습니다.</div>
      </div>
    )
  }

  const handleStatusChange = async (status: IssueStatus) => {
    // 닫힘으로 변경하는 경우 모달 표시
    if (status === 'closed') {
      setShowCompleteModal(true)
      return
    }

    try {
      const updated = await updateIssue({ status, version: issue.version })
      onUpdate?.(updated)
    } catch (err) {
      console.error('Failed to update status:', err)
    }
  }

  const handlePriorityChange = async (priority: IssuePriority) => {
    try {
      const updated = await updateIssue({ priority, version: issue.version })
      onUpdate?.(updated)
    } catch (err) {
      console.error('Failed to update priority:', err)
    }
  }

  const handleMilestoneChange = async (milestoneId: number | null) => {
    try {
      const updated = await updateIssue({ milestone_id: milestoneId, version: issue.version })
      onUpdate?.(updated)
    } catch (err) {
      console.error('Failed to update milestone:', err)
    }
  }

  const handleSaveEdit = async () => {
    try {
      const updated = await updateIssue({
        title: editTitle,
        description: editDescription,
        version: issue.version,
      })
      onUpdate?.(updated)
      setIsEditing(false)
    } catch (err) {
      console.error('Failed to update issue:', err)
    }
  }

  const handleCommentSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newComment.trim()) return

    try {
      await createComment(newComment)
      setNewComment('')
    } catch (err) {
      console.error('Failed to create comment:', err)
    }
  }

  const startEditing = () => {
    setEditTitle(issue.title)
    setEditDescription(issue.description || '')
    setIsEditing(true)
  }

  return (
    <div className="flow-issue-detail p-4">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          {onBack && (
            <button
              onClick={onBack}
              className="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 text-sm"
            >
              &larr; 뒤로
            </button>
          )}
          <span className="text-sm text-blue-600 dark:text-blue-400 font-medium">
            #{issue.issue_number}
          </span>
        </div>
        {!isEditing && (
          <button
            onClick={startEditing}
            className="px-3 py-1 text-sm text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            편집
          </button>
        )}
      </div>

      {/* Main Content */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
        {isEditing ? (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                제목
              </label>
              <input
                type="text"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                설명
              </label>
              <textarea
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleSaveEdit}
                disabled={isUpdating}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm"
              >
                {isUpdating ? '저장 중...' : '저장'}
              </button>
              <button
                onClick={() => setIsEditing(false)}
                className="px-4 py-2 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 text-sm"
              >
                취소
              </button>
            </div>
          </div>
        ) : (
          <>
            <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">{issue.title}</h1>
            {issue.description && (
              <div
                className="prose prose-sm dark:prose-invert max-w-none text-gray-700 dark:text-gray-300 mb-6"
                dangerouslySetInnerHTML={{
                  __html: issue.description_html || issue.description,
                }}
              />
            )}
          </>
        )}

        {/* Metadata */}
        <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div>
            <label className="block text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
              상태
            </label>
            <select
              value={issue.status}
              onChange={(e) => handleStatusChange(e.target.value as IssueStatus)}
              disabled={isUpdating}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            >
              <option value="open">열림</option>
              <option value="in_progress">진행 중</option>
              <option value="closed">닫힘</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
              우선순위
            </label>
            <select
              value={issue.priority}
              onChange={(e) => handlePriorityChange(e.target.value as IssuePriority)}
              disabled={isUpdating}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            >
              <option value="low">낮음</option>
              <option value="medium">보통</option>
              <option value="high">높음</option>
              <option value="urgent">긴급</option>
            </select>
          </div>
        </div>

        {/* Milestone */}
        {milestones && milestones.length > 0 && (
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
              마일스톤
            </label>
            <select
              value={issue.milestone_id || ''}
              onChange={(e) => handleMilestoneChange(e.target.value ? Number(e.target.value) : null)}
              disabled={isUpdating}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            >
              <option value="">마일스톤 없음</option>
              {milestones.map((milestone) => (
                <option key={milestone.id} value={milestone.id}>
                  {milestone.title}
                  {milestone.due_date && ` (${new Date(milestone.due_date).toLocaleDateString()})`}
                  {milestone.status === 'closed' && ' [완료]'}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Info */}
        <div className="mt-4 text-xs text-gray-500 dark:text-gray-400 space-y-1">
          <div>
            생성일: {new Date(issue.created_at).toLocaleString('ko-KR')}
          </div>
          <div>
            수정일: {new Date(issue.updated_at).toLocaleString('ko-KR')}
          </div>
        </div>
      </div>

      {/* Comments Section */}
      <div className="mt-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          댓글 {comments?.length ? `(${comments.length})` : ''}
        </h2>

        {/* Comment Form */}
        <form onSubmit={handleCommentSubmit} className="mb-4">
          <textarea
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder="댓글을 입력하세요..."
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <div className="mt-2 flex justify-end">
            <button
              type="submit"
              disabled={isCommenting || !newComment.trim()}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm"
            >
              {isCommenting ? '등록 중...' : '댓글 등록'}
            </button>
          </div>
        </form>

        {/* Comment List */}
        <div className="space-y-3">
          {comments?.map((comment) => (
            <div
              key={comment.id}
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
            >
              <div className="flex items-center gap-2 mb-2">
                <span className="font-medium text-sm text-gray-900 dark:text-gray-100">
                  {comment.user?.username || '알 수 없음'}
                </span>
                <span className="text-xs text-gray-500 dark:text-gray-400">
                  {new Date(comment.created_at).toLocaleString('ko-KR')}
                </span>
              </div>
              <div
                className="prose prose-sm dark:prose-invert max-w-none text-gray-700 dark:text-gray-300"
                dangerouslySetInnerHTML={{
                  __html: comment.content_html || comment.content,
                }}
              />
            </div>
          ))}
          {comments?.length === 0 && (
            <div className="text-center text-gray-500 dark:text-gray-400 py-8">
              아직 댓글이 없습니다.
            </div>
          )}
        </div>
      </div>

      {/* Issue Complete Modal */}
      <IssueCompleteModal
        isOpen={showCompleteModal}
        issueKey={`#${issue.issue_number}`}
        issueTitle={issue.title}
        onConfirm={handleCompleteConfirm}
        onCancel={handleCompleteCancel}
        isPending={isClosing}
      />
    </div>
  )
}
