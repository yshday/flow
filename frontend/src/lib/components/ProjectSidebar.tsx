/**
 * ProjectSidebar
 * 프로젝트 목록 사이드바 (NPM 패키지용)
 */

import type { Project } from '../types'

export interface ProjectSidebarProps {
  projects: Project[]
  selectedProjectId: number | null
  isLoading?: boolean
  onProjectClick: (project: Project) => void
  onCreateClick?: () => void
}

export function ProjectSidebar({
  projects,
  selectedProjectId,
  isLoading = false,
  onProjectClick,
  onCreateClick,
}: ProjectSidebarProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-sm" style={{ color: 'var(--flow-text-muted, #666666)' }}>
          로딩 중...
        </div>
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h2
          className="text-lg font-semibold"
          style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}
        >
          프로젝트
        </h2>
        {onCreateClick && (
          <button
            onClick={onCreateClick}
            className="p-1 rounded hover:bg-opacity-20"
            style={{
              color: '#3b82f6',
              backgroundColor: 'transparent',
            }}
            title="새 프로젝트"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        )}
      </div>

      <div className="space-y-1">
        {projects.length === 0 ? (
          <div className="text-center py-8">
            <p
              className="text-sm mb-2"
              style={{ color: 'var(--flow-text-muted, #666666)' }}
            >
              프로젝트가 없습니다
            </p>
            {onCreateClick && (
              <button
                onClick={onCreateClick}
                className="text-sm hover:underline"
                style={{ color: '#3b82f6' }}
              >
                첫 프로젝트 만들기
              </button>
            )}
          </div>
        ) : (
          projects.map((project) => (
            <button
              key={project.id}
              onClick={() => onProjectClick(project)}
              className="w-full text-left px-3 py-2 rounded-md transition-colors"
              style={{
                backgroundColor:
                  selectedProjectId === project.id
                    ? 'rgba(59, 130, 246, 0.2)'
                    : 'transparent',
                color:
                  selectedProjectId === project.id
                    ? '#60a5fa'
                    : 'var(--flow-text-secondary, #a0a0a0)',
              }}
            >
              <div className="flex items-center space-x-2">
                <span
                  className="text-xs font-mono px-1.5 py-0.5 rounded"
                  style={{
                    backgroundColor: 'var(--flow-bg-tertiary, #2a2a2a)',
                    color: 'var(--flow-text-muted, #666666)',
                  }}
                >
                  {project.key}
                </span>
                <span className="text-sm font-medium truncate">{project.name}</span>
              </div>
              {project.description && (
                <p
                  className="text-xs mt-1 truncate"
                  style={{ color: 'var(--flow-text-muted, #666666)' }}
                >
                  {project.description}
                </p>
              )}
            </button>
          ))
        )}
      </div>
    </div>
  )
}
