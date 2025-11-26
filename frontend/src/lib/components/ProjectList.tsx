/**
 * ProjectList
 * 프로젝트 목록 컴포넌트
 */

import { useFlowProjects } from '../hooks'
import type { Project } from '../types'

export interface ProjectListProps {
  onProjectClick?: (project: Project) => void
  onCreateClick?: () => void
}

export function ProjectList({ onProjectClick, onCreateClick }: ProjectListProps) {
  const { data: projects, isLoading, error } = useFlowProjects()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">프로젝트 로딩 중...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-red-500">프로젝트를 불러오는데 실패했습니다.</div>
      </div>
    )
  }

  return (
    <div className="flow-project-list" style={{ padding: '24px', paddingTop: '24px' }}>
      <div className="flex justify-between items-center" style={{ marginBottom: '24px' }}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}>프로젝트</h2>
        {onCreateClick && (
          <button
            onClick={onCreateClick}
            className="px-4 py-2 rounded-md text-sm"
            style={{ backgroundColor: '#2563eb', color: 'white' }}
          >
            새 프로젝트
          </button>
        )}
      </div>

      {projects && projects.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects.map((project) => (
            <div
              key={project.id}
              onClick={() => onProjectClick?.(project)}
              className="p-4 rounded-lg border hover:border-blue-500 hover:shadow-md transition-all cursor-pointer"
              style={{ backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)', borderColor: 'var(--flow-border, #333333)' }}
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold" style={{ color: 'var(--flow-text-primary, #f0f0f0)' }}>{project.name}</h3>
                  <p className="text-sm mt-1" style={{ color: 'var(--flow-text-muted, #666666)' }}>{project.key}</p>
                </div>
              </div>
              {project.description && (
                <p className="mt-3 text-sm line-clamp-2" style={{ color: 'var(--flow-text-secondary, #a0a0a0)' }}>
                  {project.description}
                </p>
              )}
              <div className="mt-3 text-xs" style={{ color: 'var(--flow-text-muted, #666666)' }}>
                {new Date(project.created_at).toLocaleDateString('ko-KR')}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div
          className="p-12 rounded-lg text-center"
          style={{ backgroundColor: 'var(--flow-bg-secondary, #1a1a1a)' }}
        >
          <p style={{ color: 'var(--flow-text-secondary, #a0a0a0)' }}>프로젝트가 없습니다.</p>
          {onCreateClick && (
            <button
              onClick={onCreateClick}
              className="mt-4 px-4 py-2 rounded-md text-sm"
              style={{ backgroundColor: '#2563eb', color: 'white' }}
            >
              첫 번째 프로젝트 만들기
            </button>
          )}
        </div>
      )}
    </div>
  )
}
