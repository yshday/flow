/**
 * IssueCompleteModal
 * 이슈를 완료로 이동할 때 완료 사유를 입력받는 모달
 */

import { useState } from 'react'

export interface IssueCompleteModalProps {
  isOpen: boolean
  issueKey: string // e.g., "THREAD-1"
  issueTitle: string
  onConfirm: (comment: string) => void
  onCancel: () => void
  isPending?: boolean
}

export function IssueCompleteModal({
  isOpen,
  issueKey,
  issueTitle,
  onConfirm,
  onCancel,
  isPending = false,
}: IssueCompleteModalProps) {
  const [comment, setComment] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onConfirm(comment.trim())
  }

  const handleClose = () => {
    setComment('')
    onCancel()
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
        className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 mb-4">
          <div className="flex items-center justify-center w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-full">
            <svg
              className="w-5 h-5 text-green-600 dark:text-green-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              이슈 완료
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {issueKey}: {issueTitle}
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              완료 사유 (선택)
            </label>
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="어떻게 해결했는지, 또는 완료 내용을 간략히 적어주세요..."
              rows={4}
              autoFocus
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 resize-none"
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              입력한 내용은 이슈에 댓글로 자동 등록됩니다.
            </p>
          </div>

          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={handleClose}
              disabled={isPending}
              className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md disabled:opacity-50"
            >
              취소
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="px-4 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isPending ? '처리 중...' : '완료하기'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
