/**
 * FlowIssueTracker
 * 메인 진입 컴포넌트 - 호스트 앱에 임베드되는 전체 이슈 트래커
 */

import { useEffect, useState, useCallback } from 'react'
import { FlowProvider, FlowConfig } from '../providers'
import { FlowUser, FlowCompany } from '../providers/FlowAuthProvider'
import { initFlowClient } from '../api'
import type { FlowEventCallbacks, Issue, Project } from '../types'
import { ProjectList } from './ProjectList'
import { ProjectSidebar } from './ProjectSidebar'
import { IssueTreeView } from './IssueTreeView'
import { KanbanBoard } from './KanbanBoard'
import { IssueDetail } from './IssueDetail'
import { IssueCreateModal } from './IssueCreateModal'
import { ProjectCreateModal } from './ProjectCreateModal'
import { ProjectSettings } from './ProjectSettings'
import { useFlowProjects, useFlowIssues } from '../hooks'

export interface FlowIssueTrackerProps {
  /** Flow API 설정 */
  config: FlowConfig
  /** 호스트 앱에서 주입받은 사용자 정보 */
  user: FlowUser | null
  /** 호스트 앱에서 주입받은 회사 정보 (선택) */
  company?: FlowCompany | null
  /** 호스트 앱에서 주입받은 액세스 토큰 */
  accessToken: string | null
  /** 초기 프로젝트 ID (선택) */
  initialProjectId?: number
  /** 이벤트 콜백 */
  callbacks?: FlowEventCallbacks
  /** 커스텀 클래스명 */
  className?: string
}

export type { FlowConfig, FlowUser, FlowCompany }

type View = 'projects' | 'tree' | 'board' | 'issue' | 'settings'

export function FlowIssueTracker({
  config,
  user,
  company,
  accessToken,
  initialProjectId,
  callbacks = {},
  className = '',
}: FlowIssueTrackerProps) {
  // API 클라이언트 초기화
  useEffect(() => {
    initFlowClient(config.apiBaseUrl, accessToken)
  }, [config.apiBaseUrl, accessToken])

  // Navigation state
  const [view, setView] = useState<View>(initialProjectId ? 'tree' : 'tree')
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(
    initialProjectId ?? null
  )
  const [selectedIssueId, setSelectedIssueId] = useState<number | null>(null)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [isProjectCreateModalOpen, setIsProjectCreateModalOpen] = useState(false)

  // Data fetching
  const { data: projects = [], isLoading: isLoadingProjects } = useFlowProjects()
  const { data: issues = [], isLoading: isLoadingIssues } = useFlowIssues(
    selectedProjectId || 0,
    { limit: 1000 }
  )

  // Navigation handlers
  const handleProjectClick = useCallback(
    (project: Project) => {
      setSelectedProjectId(project.id)
      setView('tree')
      callbacks.onProjectClick?.(project)
      callbacks.onNavigate?.(`/projects/${project.id}`)
    },
    [callbacks]
  )

  const handleIssueClick = useCallback(
    (issue: Issue) => {
      setSelectedIssueId(issue.id)
      setView('issue')
      callbacks.onIssueClick?.(issue)
      callbacks.onNavigate?.(`/issues/${issue.id}`)
    },
    [callbacks]
  )

  const handleBack = useCallback(() => {
    if (view === 'issue') {
      setSelectedIssueId(null)
      setView('tree')
      callbacks.onNavigate?.(`/projects/${selectedProjectId}`)
    } else if (view === 'settings') {
      setView('tree')
      callbacks.onNavigate?.(`/projects/${selectedProjectId}`)
    } else if (view === 'board') {
      setView('tree')
      callbacks.onNavigate?.(`/projects/${selectedProjectId}`)
    }
  }, [view, selectedProjectId, callbacks])

  const handleSettingsClick = useCallback(() => {
    setView('settings')
    callbacks.onNavigate?.(`/projects/${selectedProjectId}/settings`)
  }, [selectedProjectId, callbacks])

  const handleIssueCreate = useCallback(
    (issue: Issue) => {
      callbacks.onIssueCreate?.(issue)
      setIsCreateModalOpen(false)
    },
    [callbacks]
  )

  const handleIssueUpdate = useCallback(
    (issue: Issue) => {
      callbacks.onIssueUpdate?.(issue)
    },
    [callbacks]
  )

  const handleProjectCreate = useCallback(
    (project: Project) => {
      // 프로젝트 생성 후 해당 프로젝트의 트리 뷰로 이동
      setSelectedProjectId(project.id)
      setView('tree')
      setIsProjectCreateModalOpen(false)
      callbacks.onNavigate?.(`/projects/${project.id}`)
    },
    [callbacks]
  )

  // 인증되지 않은 경우
  if (!accessToken || !user) {
    return (
      <div
        className={`flow-issue-tracker relative h-full min-h-full ${className}`}
        style={{ backgroundColor: 'var(--flow-bg-primary, #101010)' }}
      >
        <div
          className="flex items-center justify-center h-64 rounded-lg"
          style={{ backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)' }}
        >
          <p style={{ color: 'var(--flow-text-muted, #666666)' }}>로그인이 필요합니다.</p>
        </div>
      </div>
    )
  }

  return (
    <FlowProvider
      config={config}
      user={user}
      company={company}
      accessToken={accessToken}
      callbacks={{
        ...callbacks,
        onIssueCreate: handleIssueCreate,
        onIssueUpdate: handleIssueUpdate,
      }}
    >
      <div
        className={`flow-issue-tracker relative h-full min-h-screen flex overflow-hidden ${className}`}
        style={{ backgroundColor: 'var(--flow-bg-primary, #101010)' }}
      >
        {/* Sidebar */}
        <div
          className="w-52 flex-shrink-0 border-r overflow-y-auto"
          style={{
            backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)',
            borderColor: 'var(--flow-border, #333333)',
          }}
        >
          <ProjectSidebar
            projects={projects}
            selectedProjectId={selectedProjectId}
            isLoading={isLoadingProjects}
            onProjectClick={handleProjectClick}
            onCreateClick={() => setIsProjectCreateModalOpen(true)}
          />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto">
          {!selectedProjectId && (
            <div className="flex items-center justify-center h-full">
              <p style={{ color: 'var(--flow-text-muted, #666666)' }}>
                프로젝트를 선택하세요
              </p>
            </div>
          )}

          {selectedProjectId && view === 'tree' && (
            <IssueTreeView
              projectId={selectedProjectId}
              issues={issues}
              isLoading={isLoadingIssues}
              onIssueClick={handleIssueClick}
              onCreateClick={() => setIsCreateModalOpen(true)}
            />
          )}

          {view === 'board' && selectedProjectId && (
            <KanbanBoard
              projectId={selectedProjectId}
              onBack={handleBack}
              onIssueClick={handleIssueClick}
              onCreateClick={() => setIsCreateModalOpen(true)}
              onSettingsClick={handleSettingsClick}
            />
          )}

          {view === 'issue' && selectedIssueId && (
            <IssueDetail
              issueId={selectedIssueId}
              onBack={handleBack}
              onUpdate={handleIssueUpdate}
            />
          )}

          {view === 'settings' && selectedProjectId && (
            <ProjectSettings
              projectId={selectedProjectId}
              onBack={handleBack}
              currentUserId={user?.id}
            />
          )}
        </div>

        {/* Issue Create Modal */}
        {selectedProjectId && (
          <IssueCreateModal
            projectId={selectedProjectId}
            isOpen={isCreateModalOpen}
            onClose={() => setIsCreateModalOpen(false)}
            onCreated={handleIssueCreate}
          />
        )}

        {/* Project Create Modal */}
        <ProjectCreateModal
          isOpen={isProjectCreateModalOpen}
          onClose={() => setIsProjectCreateModalOpen(false)}
          onCreated={handleProjectCreate}
        />
      </div>
    </FlowProvider>
  )
}
