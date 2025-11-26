/**
 * ProjectCreateModal
 * 새 프로젝트 생성 모달
 */

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { getFlowClient } from '../api'
import type { Project } from '../types'

export interface ProjectCreateModalProps {
  isOpen: boolean
  onClose: () => void
  onCreated?: (project: Project) => void
}

export function ProjectCreateModal({ isOpen, onClose, onCreated }: ProjectCreateModalProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [key, setKey] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState<string | null>(null)

  const createMutation = useMutation({
    mutationFn: async (data: { name: string; key: string; description?: string }) => {
      const client = getFlowClient()
      const response = await client.post<Project>('/projects', data)
      return response.data
    },
    onSuccess: (project) => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects'] })
      onCreated?.(project)
      handleClose()
    },
    onError: (err: any) => {
      const errorData = err.response?.data?.error
      if (typeof errorData === 'string') {
        setError(errorData)
      } else if (errorData?.message) {
        setError(errorData.message)
      } else {
        setError('프로젝트 생성에 실패했습니다.')
      }
    },
  })

  const handleClose = () => {
    setName('')
    setKey('')
    setDescription('')
    setError(null)
    onClose()
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      setError('프로젝트 이름을 입력해주세요.')
      return
    }
    if (!key.trim()) {
      setError('프로젝트 키를 입력해주세요.')
      return
    }
    createMutation.mutate({
      name: name.trim(),
      key: key.trim().toUpperCase(),
      description: description.trim() || undefined,
    })
  }

  // 이름 변경 시 자동으로 키 생성
  const handleNameChange = (value: string) => {
    setName(value)
    if (!key || key === generateKey(name)) {
      setKey(generateKey(value))
    }
  }

  const generateKey = (name: string): string => {
    return name
      .trim()
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, '')
      .slice(0, 10)
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
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          새 프로젝트 만들기
        </h3>

        <form onSubmit={handleSubmit}>
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md text-sm">
              {error}
            </div>
          )}

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              프로젝트 이름 *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="내 프로젝트"
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              프로젝트 키 *
            </label>
            <input
              type="text"
              value={key}
              onChange={(e) => setKey(e.target.value.toUpperCase())}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="PROJ"
              maxLength={10}
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              이슈 번호에 사용됩니다 (예: PROJ-1, PROJ-2)
            </p>
          </div>

          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              설명 (선택)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              rows={3}
              placeholder="프로젝트에 대한 설명"
            />
          </div>

          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md"
            >
              취소
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {createMutation.isPending ? '생성 중...' : '프로젝트 만들기'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
