import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type { Project } from '../../types';
import LoadingSpinner from '../common/LoadingSpinner';
import { useState } from 'react';
import CreateProjectModal from '../common/CreateProjectModal';

export default function ProjectSidebar() {
  const navigate = useNavigate();
  const { id: currentProjectId } = useParams<{ id: string }>();
  const [showCreateModal, setShowCreateModal] = useState(false);

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await apiClient.get<Project[]>('/projects');
      return response.data;
    },
  });

  const handleProjectClick = (projectId: number) => {
    navigate(`/projects/${projectId}`);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="sm" />
      </div>
    );
  }

  return (
    <>
      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">프로젝트</h2>
          <button
            onClick={() => setShowCreateModal(true)}
            className="p-1 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded"
            title="새 프로젝트"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>

        <div className="space-y-1">
          {projects.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400 text-sm mb-2">프로젝트가 없습니다</p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="text-blue-600 dark:text-blue-400 text-sm hover:underline"
              >
                첫 프로젝트 만들기
              </button>
            </div>
          ) : (
            projects.map((project) => (
              <button
                key={project.id}
                onClick={() => handleProjectClick(project.id)}
                className={`w-full text-left px-3 py-2 rounded-md transition-colors ${
                  currentProjectId === project.id.toString()
                    ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
              >
                <div className="flex items-center space-x-2">
                  <span className="text-xs font-mono bg-gray-200 dark:bg-gray-700 px-1.5 py-0.5 rounded">
                    {project.key}
                  </span>
                  <span className="text-sm font-medium truncate">{project.name}</span>
                </div>
                {project.description && (
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">
                    {project.description}
                  </p>
                )}
              </button>
            ))
          )}
        </div>
      </div>

      <CreateProjectModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={() => {
          setShowCreateModal(false);
        }}
      />
    </>
  );
}
