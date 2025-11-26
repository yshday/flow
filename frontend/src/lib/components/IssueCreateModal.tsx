/**
 * IssueCreateModal
 * 이슈 생성 모달 컴포넌트
 */

import { useState, useEffect } from 'react'
import { useFlowCreateIssue, useFlowLabels, useFlowBoardColumns, useFlowMilestones, useFlowEpics } from '../hooks'
import type { Issue, IssuePriority, IssueType } from '../types'

export interface IssueCreateModalProps {
  projectId: number
  isOpen: boolean
  onClose: () => void
  onCreated?: (issue: Issue) => void
}

export function IssueCreateModal({
  projectId,
  isOpen,
  onClose,
  onCreated,
}: IssueCreateModalProps) {
  const { mutateAsync: createIssue, isPending } = useFlowCreateIssue(projectId)
  const { data: labels } = useFlowLabels(projectId)
  const { data: columns } = useFlowBoardColumns(projectId)
  const { data: milestones } = useFlowMilestones(projectId)
  const { data: epics = [] } = useFlowEpics(projectId)

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState<IssuePriority>('medium')
  const [issueType, setIssueType] = useState<IssueType>('task')
  const [columnId, setColumnId] = useState<number | undefined>(undefined)
  const [milestoneId, setMilestoneId] = useState<number | undefined>(undefined)
  const [epicId, setEpicId] = useState<number | undefined>(undefined)
  const [selectedLabelIds, setSelectedLabelIds] = useState<number[]>([])

  // 기본적으로 첫 번째 컬럼(Backlog) 선택
  useEffect(() => {
    if (columns && columns.length > 0 && columnId === undefined) {
      const sortedColumns = [...columns].sort((a, b) => a.position - b.position)
      setColumnId(sortedColumns[0].id)
    }
  }, [columns, columnId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return

    try {
      const issue = await createIssue({
        title: title.trim(),
        description: description.trim() || undefined,
        priority,
        issue_type: issueType,
        column_id: columnId,
        milestone_id: milestoneId,
        epic_id: epicId,
        label_ids: selectedLabelIds.length > 0 ? selectedLabelIds : undefined,
      })
      onCreated?.(issue)
      resetForm()
      onClose()
    } catch (err) {
      console.error('Failed to create issue:', err)
    }
  }

  const resetForm = () => {
    setTitle('')
    setDescription('')
    setPriority('medium')
    setIssueType('task')
    // 첫 번째 컬럼으로 리셋
    if (columns && columns.length > 0) {
      const sortedColumns = [...columns].sort((a, b) => a.position - b.position)
      setColumnId(sortedColumns[0].id)
    }
    setMilestoneId(undefined)
    setEpicId(undefined)
    setSelectedLabelIds([])
  }

  const handleClose = () => {
    resetForm()
    onClose()
  }

  const toggleLabel = (labelId: number) => {
    setSelectedLabelIds((prev) =>
      prev.includes(labelId) ? prev.filter((id) => id !== labelId) : [...prev, labelId]
    )
  }

  if (!isOpen) return null

  return (
    <div
      className="absolute inset-0 z-50 flex items-center justify-center overflow-hidden"
      style={{
        backgroundColor: 'rgba(0, 0, 0, 0.7)',
        minHeight: '100%',
      }}
      onClick={handleClose}
    >
      {/* Modal */}
      <div
        className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">새 이슈</h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Title */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                제목 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="이슈 제목을 입력하세요"
                required
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                설명
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="이슈에 대한 설명을 입력하세요"
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Type and Priority */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  유형
                </label>
                <select
                  value={issueType}
                  onChange={(e) => setIssueType(e.target.value as IssueType)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="task">📋 작업</option>
                  <option value="bug">🐛 버그</option>
                  <option value="feature">✨ 기능</option>
                  <option value="improvement">⚡ 개선</option>
                  <option value="epic">🎯 에픽</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  우선순위
                </label>
                <select
                  value={priority}
                  onChange={(e) => setPriority(e.target.value as IssuePriority)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="low">낮음</option>
                  <option value="medium">보통</option>
                  <option value="high">높음</option>
                  <option value="urgent">긴급</option>
                </select>
              </div>
            </div>

            {/* Column */}
            {columns && columns.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  컬럼
                </label>
                <select
                  value={columnId || ''}
                  onChange={(e) => setColumnId(Number(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {[...columns].sort((a, b) => a.position - b.position).map((column) => (
                    <option key={column.id} value={column.id}>
                      {column.name}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Milestone */}
            {milestones && milestones.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  마일스톤
                </label>
                <select
                  value={milestoneId || ''}
                  onChange={(e) => setMilestoneId(e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">마일스톤 없음</option>
                  {milestones.filter(m => m.status === 'open').map((milestone) => (
                    <option key={milestone.id} value={milestone.id}>
                      {milestone.title}
                      {milestone.due_date && ` (${new Date(milestone.due_date).toLocaleDateString()})`}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Epic Selection - Only show if issue type is NOT epic */}
            {issueType !== 'epic' && epics.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  에픽
                </label>
                <select
                  value={epicId || ''}
                  onChange={(e) => setEpicId(e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">에픽 없음</option>
                  {epics.map((epic) => (
                    <option key={epic.id} value={epic.id}>
                      🎯 {epic.project_id}-{epic.issue_number} {epic.title}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Labels */}
            {labels && labels.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  라벨
                </label>
                <div className="flex flex-wrap gap-2">
                  {labels.map((label) => (
                    <button
                      key={label.id}
                      type="button"
                      onClick={() => toggleLabel(label.id)}
                      className={`px-3 py-1 text-sm rounded-full border transition-colors ${
                        selectedLabelIds.includes(label.id)
                          ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                          : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                      }`}
                      style={{
                        backgroundColor: selectedLabelIds.includes(label.id)
                          ? `${label.color}20`
                          : undefined,
                        borderColor: selectedLabelIds.includes(label.id)
                          ? label.color
                          : undefined,
                        color: selectedLabelIds.includes(label.id) ? label.color : undefined,
                      }}
                    >
                      {label.name}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                type="button"
                onClick={handleClose}
                className="px-4 py-2 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                취소
              </button>
              <button
                type="submit"
                disabled={isPending || !title.trim()}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {isPending ? '생성 중...' : '이슈 생성'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
