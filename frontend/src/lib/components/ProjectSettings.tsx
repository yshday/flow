/**
 * ProjectSettings
 * 프로젝트 설정 컴포넌트 - 라벨, 마일스톤, 멤버 관리
 */

import { useState, useEffect, useCallback } from 'react'
import {
  useFlowProject,
  useFlowLabels,
  useFlowCreateLabel,
  useFlowUpdateLabel,
  useFlowDeleteLabel,
  useFlowMilestones,
  useFlowCreateMilestone,
  useFlowUpdateMilestone,
  useFlowDeleteMilestone,
  useFlowProjectMembers,
  useFlowAddMember,
  useFlowUpdateMemberRole,
  useFlowRemoveMember,
  useFlowSearchUsers,
} from '../hooks'
import { useFlowCallbacks } from '../providers'
import type { Label, Milestone, ProjectMember, ProjectRole, User, MemberSearchResult } from '../types'

export interface ProjectSettingsProps {
  projectId: number
  onBack?: () => void
  currentUserId?: number
}

type SettingsTab = 'labels' | 'milestones' | 'members'

export function ProjectSettings({ projectId, onBack, currentUserId }: ProjectSettingsProps) {
  const callbacks = useFlowCallbacks()
  const { data: project, isLoading: projectLoading } = useFlowProject(projectId)
  const { data: labels, isLoading: labelsLoading } = useFlowLabels(projectId)
  const { data: milestones, isLoading: milestonesLoading } = useFlowMilestones(projectId)
  const { data: members, isLoading: membersLoading } = useFlowProjectMembers(projectId)

  const { mutateAsync: createLabel } = useFlowCreateLabel(projectId)
  const { mutateAsync: updateLabel } = useFlowUpdateLabel(projectId)
  const { mutateAsync: deleteLabel } = useFlowDeleteLabel(projectId)

  const { mutateAsync: createMilestone } = useFlowCreateMilestone(projectId)
  const { mutateAsync: updateMilestone } = useFlowUpdateMilestone(projectId)
  const { mutateAsync: deleteMilestone } = useFlowDeleteMilestone(projectId)

  const { mutateAsync: addMember } = useFlowAddMember(projectId)
  const { mutateAsync: updateMemberRole } = useFlowUpdateMemberRole(projectId)
  const { mutateAsync: removeMember } = useFlowRemoveMember(projectId)

  const [activeTab, setActiveTab] = useState<SettingsTab>('labels')

  // Label Modal State
  const [isLabelModalOpen, setIsLabelModalOpen] = useState(false)
  const [editingLabel, setEditingLabel] = useState<Label | null>(null)
  const [labelName, setLabelName] = useState('')
  const [labelColor, setLabelColor] = useState('#3B82F6')
  const [labelDescription, setLabelDescription] = useState('')

  // Milestone Modal State
  const [isMilestoneModalOpen, setIsMilestoneModalOpen] = useState(false)
  const [editingMilestone, setEditingMilestone] = useState<Milestone | null>(null)
  const [milestoneTitle, setMilestoneTitle] = useState('')
  const [milestoneDescription, setMilestoneDescription] = useState('')
  const [milestoneDueDate, setMilestoneDueDate] = useState('')

  // Member Modal State
  const [isMemberModalOpen, setIsMemberModalOpen] = useState(false)
  const [userSearchQuery, setUserSearchQuery] = useState('')
  const [selectedUsers, setSelectedUsers] = useState<(User | MemberSearchResult)[]>([])
  const [newMemberRole, setNewMemberRole] = useState<ProjectRole>('member')

  // 호스트 앱 콜백 검색 결과
  const [externalSearchResults, setExternalSearchResults] = useState<MemberSearchResult[]>([])
  const [isSearching, setIsSearching] = useState(false)

  // Flow API 검색 (콜백이 없을 때 폴백)
  const { data: flowSearchResults } = useFlowSearchUsers(
    callbacks.onSearchMembers ? '' : userSearchQuery // 콜백이 있으면 Flow API 비활성화
  )

  // 호스트 앱 콜백으로 멤버 검색
  useEffect(() => {
    if (!callbacks.onSearchMembers || userSearchQuery.length < 2) {
      setExternalSearchResults([])
      return
    }

    const searchTimeout = setTimeout(async () => {
      setIsSearching(true)
      try {
        const results = await callbacks.onSearchMembers!(userSearchQuery)
        setExternalSearchResults(results)
      } catch (err) {
        console.error('Failed to search members:', err)
        setExternalSearchResults([])
      } finally {
        setIsSearching(false)
      }
    }, 300) // 300ms debounce

    return () => clearTimeout(searchTimeout)
  }, [userSearchQuery, callbacks.onSearchMembers])

  // 검색 결과 통합 (콜백 결과 우선)
  const searchResults = callbacks.onSearchMembers ? externalSearchResults : flowSearchResults

  // Filter out existing members from search results
  const filteredSearchResults = searchResults?.filter(
    (user) => !members?.some((member) => member.user_id === user.id) &&
              !selectedUsers.some((u) => u.id === user.id)
  )

  const isCurrentUserOwner = members?.some(
    (member) => member.user_id === currentUserId && member.role === 'owner'
  )

  if (projectLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500 dark:text-gray-400">로딩 중...</div>
      </div>
    )
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-red-500 dark:text-red-400">프로젝트를 찾을 수 없습니다.</div>
      </div>
    )
  }

  // === Label Handlers ===
  const openLabelModal = (label?: Label) => {
    if (label) {
      setEditingLabel(label)
      setLabelName(label.name)
      setLabelColor(label.color)
      setLabelDescription(label.description || '')
    } else {
      setEditingLabel(null)
      setLabelName('')
      setLabelColor('#3B82F6')
      setLabelDescription('')
    }
    setIsLabelModalOpen(true)
  }

  const closeLabelModal = () => {
    setIsLabelModalOpen(false)
    setEditingLabel(null)
    setLabelName('')
    setLabelColor('#3B82F6')
    setLabelDescription('')
  }

  const handleLabelSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!labelName.trim()) return

    try {
      if (editingLabel) {
        await updateLabel({
          labelId: editingLabel.id,
          data: { name: labelName, color: labelColor, description: labelDescription || undefined },
        })
      } else {
        await createLabel({
          name: labelName,
          color: labelColor,
          description: labelDescription || undefined,
        })
      }
      closeLabelModal()
    } catch (err) {
      console.error('Failed to save label:', err)
    }
  }

  const handleDeleteLabel = async (labelId: number) => {
    if (!window.confirm('이 라벨을 삭제하시겠습니까?')) return
    try {
      await deleteLabel(labelId)
    } catch (err) {
      console.error('Failed to delete label:', err)
    }
  }

  // === Milestone Handlers ===
  const openMilestoneModal = (milestone?: Milestone) => {
    if (milestone) {
      setEditingMilestone(milestone)
      setMilestoneTitle(milestone.title)
      setMilestoneDescription(milestone.description || '')
      setMilestoneDueDate(milestone.due_date?.split('T')[0] || '')
    } else {
      setEditingMilestone(null)
      setMilestoneTitle('')
      setMilestoneDescription('')
      setMilestoneDueDate('')
    }
    setIsMilestoneModalOpen(true)
  }

  const closeMilestoneModal = () => {
    setIsMilestoneModalOpen(false)
    setEditingMilestone(null)
    setMilestoneTitle('')
    setMilestoneDescription('')
    setMilestoneDueDate('')
  }

  const handleMilestoneSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!milestoneTitle.trim()) return

    // Convert date to RFC3339 format for backend
    const formattedDueDate = milestoneDueDate ? `${milestoneDueDate}T00:00:00Z` : undefined

    try {
      if (editingMilestone) {
        await updateMilestone({
          milestoneId: editingMilestone.id,
          data: {
            title: milestoneTitle,
            description: milestoneDescription || undefined,
            due_date: formattedDueDate,
          },
        })
      } else {
        await createMilestone({
          title: milestoneTitle,
          description: milestoneDescription || undefined,
          due_date: formattedDueDate,
        })
      }
      closeMilestoneModal()
    } catch (err) {
      console.error('Failed to save milestone:', err)
    }
  }

  const handleDeleteMilestone = async (milestoneId: number) => {
    if (!window.confirm('이 마일스톤을 삭제하시겠습니까?')) return
    try {
      await deleteMilestone(milestoneId)
    } catch (err) {
      console.error('Failed to delete milestone:', err)
    }
  }

  // === Member Handlers ===
  const openMemberModal = () => {
    setUserSearchQuery('')
    setSelectedUsers([])
    setNewMemberRole('member')
    setIsMemberModalOpen(true)
  }

  const closeMemberModal = () => {
    setIsMemberModalOpen(false)
    setUserSearchQuery('')
    setSelectedUsers([])
    setNewMemberRole('member')
  }

  const handleAddMembers = async (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedUsers.length === 0) return

    try {
      for (const user of selectedUsers) {
        await addMember({ user_id: user.id, role: newMemberRole })
      }
      closeMemberModal()
    } catch (err) {
      console.error('Failed to add members:', err)
    }
  }

  const handleRoleChange = async (userId: number, role: ProjectRole) => {
    try {
      await updateMemberRole({ userId, role })
    } catch (err) {
      console.error('Failed to update role:', err)
    }
  }

  const handleRemoveMember = async (userId: number) => {
    if (!window.confirm('이 멤버를 제거하시겠습니까?')) return
    try {
      await removeMember(userId)
    } catch (err) {
      console.error('Failed to remove member:', err)
    }
  }

  const getRoleLabel = (role: ProjectRole): string => {
    const labels: Record<ProjectRole, string> = {
      owner: '소유자',
      admin: '관리자',
      member: '멤버',
      viewer: '뷰어',
    }
    return labels[role]
  }

  const colorPresets = [
    '#EF4444', '#F97316', '#F59E0B', '#EAB308', '#84CC16',
    '#22C55E', '#10B981', '#14B8A6', '#06B6D4', '#0EA5E9',
    '#3B82F6', '#6366F1', '#8B5CF6', '#A855F7', '#D946EF',
    '#EC4899', '#F43F5E', '#6B7280', '#374151', '#1F2937',
  ]

  return (
    <div className="flow-project-settings p-4">
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
          <div>
            <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">프로젝트 설정</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400">{project.name}</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700 mb-10">
        <nav className="flex gap-6 pt-4 pb-1">
          {(['labels', 'milestones', 'members'] as SettingsTab[]).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              {tab === 'labels' ? '라벨' : tab === 'milestones' ? '마일스톤' : '멤버'}
            </button>
          ))}
        </nav>
      </div>

      {/* Labels Tab */}
      {activeTab === 'labels' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">라벨</h2>
            <button
              onClick={() => openLabelModal()}
              className="px-3 py-1.5 text-sm rounded-md"
              style={{ backgroundColor: '#2563eb', color: 'white' }}
            >
              새 라벨
            </button>
          </div>

          {labelsLoading ? (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">로딩 중...</div>
          ) : labels && labels.length > 0 ? (
            <div className="space-y-2">
              {labels.map((label) => (
                <div
                  key={label.id}
                  className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <div className="flex items-center gap-3">
                    <span
                      className="px-3 py-1 text-sm font-medium rounded"
                      style={{
                        backgroundColor: label.color + '20',
                        color: label.color,
                        border: `1px solid ${label.color}`,
                      }}
                    >
                      {label.name}
                    </span>
                    {label.description && (
                      <span className="text-sm text-gray-600 dark:text-gray-400">{label.description}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => openLabelModal(label)}
                      className="text-gray-500 hover:text-blue-600 dark:hover:text-blue-400"
                    >
                      수정
                    </button>
                    <button
                      onClick={() => handleDeleteLabel(label.id)}
                      className="text-gray-500 hover:text-red-600 dark:hover:text-red-400"
                    >
                      삭제
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">
              라벨이 없습니다. 새 라벨을 만들어보세요.
            </div>
          )}
        </div>
      )}

      {/* Milestones Tab */}
      {activeTab === 'milestones' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">마일스톤</h2>
            <button
              onClick={() => openMilestoneModal()}
              className="px-3 py-1.5 text-sm rounded-md"
              style={{ backgroundColor: '#2563eb', color: 'white' }}
            >
              새 마일스톤
            </button>
          </div>

          {milestonesLoading ? (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">로딩 중...</div>
          ) : milestones && milestones.length > 0 ? (
            <div className="space-y-2">
              {milestones.map((milestone) => (
                <div
                  key={milestone.id}
                  className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-medium text-gray-900 dark:text-gray-100">{milestone.title}</h3>
                      <span
                        className={`px-2 py-0.5 text-xs font-medium rounded ${
                          milestone.status === 'open'
                            ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400'
                        }`}
                      >
                        {milestone.status === 'open' ? '진행 중' : '완료'}
                      </span>
                    </div>
                    {milestone.description && (
                      <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">{milestone.description}</p>
                    )}
                    {milestone.due_date && (
                      <p className="text-xs text-gray-500 dark:text-gray-500">
                        마감일: {new Date(milestone.due_date).toLocaleDateString('ko-KR')}
                      </p>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => openMilestoneModal(milestone)}
                      className="text-gray-500 hover:text-blue-600 dark:hover:text-blue-400"
                    >
                      수정
                    </button>
                    <button
                      onClick={() => handleDeleteMilestone(milestone.id)}
                      className="text-gray-500 hover:text-red-600 dark:hover:text-red-400"
                    >
                      삭제
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">
              마일스톤이 없습니다. 새 마일스톤을 만들어보세요.
            </div>
          )}
        </div>
      )}

      {/* Members Tab */}
      {activeTab === 'members' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">멤버</h2>
            <button
              onClick={openMemberModal}
              className="px-3 py-1.5 text-sm rounded-md"
              style={{ backgroundColor: '#2563eb', color: 'white' }}
            >
              멤버 추가
            </button>
          </div>

          {membersLoading ? (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">로딩 중...</div>
          ) : members && members.length > 0 ? (
            <div className="space-y-2">
              {members.map((member) => (
                <div
                  key={member.user_id}
                  className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                      {member.user?.username?.[0]?.toUpperCase() || '?'}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900 dark:text-gray-100">
                          {member.user?.username || member.user?.email}
                        </span>
                        {member.user_id === currentUserId && (
                          <span className="text-xs text-gray-500 dark:text-gray-400">(나)</span>
                        )}
                      </div>
                      {member.user?.email && member.user?.username && (
                        <span className="text-sm text-gray-500 dark:text-gray-400">{member.user.email}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <select
                      value={member.role}
                      onChange={(e) => handleRoleChange(member.user_id, e.target.value as ProjectRole)}
                      disabled={member.user_id === currentUserId || (!isCurrentUserOwner && member.role === 'owner')}
                      className="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <option value="owner" disabled={!isCurrentUserOwner}>소유자</option>
                      <option value="admin">관리자</option>
                      <option value="member">멤버</option>
                      <option value="viewer">뷰어</option>
                    </select>
                    {member.role !== 'owner' && member.user_id !== currentUserId && (
                      <button
                        onClick={() => handleRemoveMember(member.user_id)}
                        className="text-gray-500 hover:text-red-600 dark:hover:text-red-400"
                      >
                        제거
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-gray-500 dark:text-gray-400 text-center py-8">
              프로젝트 멤버가 없습니다.
            </div>
          )}
        </div>
      )}

      {/* Label Modal */}
      {isLabelModalOpen && (
        <div
          className="absolute inset-0 z-50 flex items-center justify-center overflow-hidden"
          style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', minHeight: '100%' }}
          onClick={closeLabelModal}
        >
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
              {editingLabel ? '라벨 수정' : '새 라벨'}
            </h3>
            <form onSubmit={handleLabelSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    이름 *
                  </label>
                  <input
                    type="text"
                    value={labelName}
                    onChange={(e) => setLabelName(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    색상
                  </label>
                  <div className="flex flex-wrap gap-2 mb-2">
                    {colorPresets.map((color) => (
                      <button
                        key={color}
                        type="button"
                        onClick={() => setLabelColor(color)}
                        className={`w-6 h-6 rounded-full border-2 ${
                          labelColor === color ? 'border-gray-900 dark:border-white' : 'border-transparent'
                        }`}
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </div>
                  <input
                    type="color"
                    value={labelColor}
                    onChange={(e) => setLabelColor(e.target.value)}
                    className="w-full h-10 cursor-pointer"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    설명
                  </label>
                  <input
                    type="text"
                    value={labelDescription}
                    onChange={(e) => setLabelDescription(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={closeLabelModal}
                  className="px-4 py-2 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  취소
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {editingLabel ? '수정' : '생성'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Milestone Modal */}
      {isMilestoneModalOpen && (
        <div
          className="absolute inset-0 z-50 flex items-center justify-center overflow-hidden"
          style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', minHeight: '100%' }}
          onClick={closeMilestoneModal}
        >
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
              {editingMilestone ? '마일스톤 수정' : '새 마일스톤'}
            </h3>
            <form onSubmit={handleMilestoneSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    제목 *
                  </label>
                  <input
                    type="text"
                    value={milestoneTitle}
                    onChange={(e) => setMilestoneTitle(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    설명
                  </label>
                  <textarea
                    value={milestoneDescription}
                    onChange={(e) => setMilestoneDescription(e.target.value)}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    마감일
                  </label>
                  <input
                    type="date"
                    value={milestoneDueDate}
                    onChange={(e) => setMilestoneDueDate(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={closeMilestoneModal}
                  className="px-4 py-2 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  취소
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {editingMilestone ? '수정' : '생성'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Member Modal */}
      {isMemberModalOpen && (
        <div
          className="absolute inset-0 z-50 flex items-center justify-center overflow-hidden"
          style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', minHeight: '100%' }}
          onClick={closeMemberModal}
        >
          <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">멤버 추가</h3>
            <form onSubmit={handleAddMembers}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    사용자 검색
                  </label>
                  {selectedUsers.length > 0 && (
                    <div className="flex flex-wrap gap-2 mb-2">
                      {selectedUsers.map((user) => (
                        <div
                          key={user.id}
                          className="flex items-center gap-1 px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 rounded-md text-sm"
                        >
                          <span>{'name' in user ? user.name : user.username}</span>
                          <button
                            type="button"
                            onClick={() => setSelectedUsers(selectedUsers.filter((u) => u.id !== user.id))}
                            className="hover:bg-blue-200 dark:hover:bg-blue-800 rounded-full p-0.5"
                          >
                            &times;
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                  <input
                    type="text"
                    value={userSearchQuery}
                    onChange={(e) => setUserSearchQuery(e.target.value)}
                    placeholder="이메일 또는 사용자명"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  />
                  {userSearchQuery.length >= 2 && (
                    <div
                      className="mt-2 border border-gray-300 dark:border-gray-600 rounded-md overflow-y-auto"
                      style={{ maxHeight: '200px' }}
                    >
                      {isSearching ? (
                        <div className="px-3 py-2 text-center text-gray-500 dark:text-gray-400">검색 중...</div>
                      ) : filteredSearchResults && filteredSearchResults.length > 0 ? (
                        filteredSearchResults.slice(0, 50).map((user) => {
                          const displayName = 'name' in user ? user.name : user.username
                          const avatarUrl = 'avatar_url' in user ? user.avatar_url : undefined
                          const initial = displayName?.charAt(0).toUpperCase() || '?'

                          return (
                            <button
                              key={user.id}
                              type="button"
                              onClick={() => setSelectedUsers([...selectedUsers, user])}
                              className="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 border-b border-gray-200 dark:border-gray-600 last:border-b-0 flex items-center gap-3"
                            >
                              {/* 프로필 이미지 또는 이니셜 */}
                              <div className="flex-shrink-0">
                                {avatarUrl ? (
                                  <img
                                    src={avatarUrl}
                                    alt={displayName}
                                    className="w-8 h-8 rounded-full object-cover"
                                  />
                                ) : (
                                  <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm font-medium">
                                    {initial}
                                  </div>
                                )}
                              </div>
                              <div className="flex-1 min-w-0">
                                <div className="font-medium text-gray-900 dark:text-gray-100">
                                  {displayName}
                                  {'rank' in user && user.rank && (
                                    <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">{user.rank}</span>
                                  )}
                                </div>
                                <div className="text-sm text-gray-500 dark:text-gray-400 truncate">
                                  {user.email}
                                  {'department' in user && user.department && (
                                    <span className="ml-2">{user.department}</span>
                                  )}
                                </div>
                              </div>
                            </button>
                          )
                        })
                      ) : (
                        <div className="px-3 py-2 text-center text-gray-500 dark:text-gray-400">검색 결과가 없습니다.</div>
                      )}
                      {filteredSearchResults && filteredSearchResults.length > 50 && (
                        <div className="px-3 py-2 text-center text-gray-500 dark:text-gray-400 text-sm bg-gray-50 dark:bg-gray-700">
                          +{filteredSearchResults.length - 50}명 더 있습니다. 검색어를 더 구체적으로 입력하세요.
                        </div>
                      )}
                    </div>
                  )}
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    역할
                  </label>
                  <select
                    value={newMemberRole}
                    onChange={(e) => setNewMemberRole(e.target.value as ProjectRole)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  >
                    <option value="admin">관리자</option>
                    <option value="member">멤버</option>
                    <option value="viewer">뷰어</option>
                  </select>
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={closeMemberModal}
                  className="px-4 py-2 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  취소
                </button>
                <button
                  type="submit"
                  disabled={selectedUsers.length === 0}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  추가
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
