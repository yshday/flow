import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type { Issue } from '../../types';
import LoadingSpinner from '../common/LoadingSpinner';
import IssueTreeItem from './IssueTreeItem';
import { useState } from 'react';
import CreateIssueModal from '../issue/CreateIssueModal';

interface IssueTreeViewProps {
  projectId: number;
}

export default function IssueTreeView({ projectId }: IssueTreeViewProps) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [filterPriority, setFilterPriority] = useState<string>('');

  const { data: issues = [], isLoading } = useQuery({
    queryKey: ['issues', projectId, searchQuery, filterStatus, filterPriority],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.append('limit', '1000'); // Get all issues
      if (searchQuery) params.append('search', searchQuery);
      if (filterStatus) params.append('status', filterStatus);
      if (filterPriority) params.append('priority', filterPriority);

      const response = await apiClient.get<Issue[]>(`/projects/${projectId}/issues?${params}`);
      return response.data;
    },
    enabled: !!projectId,
  });

  // Build issue tree from flat list
  const issuesWithChildren = issues.map((issue) => {
    // For epics, populate epic_issues array with child issues
    if (issue.issue_type === 'epic') {
      const epicIssues = issues.filter((i) => i.epic_id === issue.id);
      return { ...issue, epic_issues: epicIssues };
    }

    // For regular issues, populate subtasks array
    const subtasks = issues.filter((i) => i.parent_issue_id === issue.id);
    return { ...issue, subtasks };
  });

  // Build issue tree: only show root-level issues (no parent_issue_id and not in an epic)
  const rootIssues = issuesWithChildren.filter((issue) => !issue.parent_issue_id && !issue.epic_id);

  // Sort issues: pinned first, then by issue number
  const sortedRootIssues = rootIssues.sort((a, b) => {
    if (a.is_pinned && !b.is_pinned) return -1;
    if (!a.is_pinned && b.is_pinned) return 1;
    return a.issue_number - b.issue_number;
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  return (
    <>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">이슈</h2>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            새 이슈
          </button>
        </div>

        {/* Filters */}
        <div className="mb-4 flex flex-wrap gap-4">
          <input
            type="text"
            placeholder="검색..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="flex-1 min-w-[200px] px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          />

          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          >
            <option value="">모든 상태</option>
            <option value="open">열림</option>
            <option value="in_progress">진행 중</option>
            <option value="closed">닫힘</option>
          </select>

          <select
            value={filterPriority}
            onChange={(e) => setFilterPriority(e.target.value)}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          >
            <option value="">모든 우선순위</option>
            <option value="low">낮음</option>
            <option value="medium">보통</option>
            <option value="high">높음</option>
            <option value="urgent">긴급</option>
          </select>
        </div>

        {/* Issue Tree */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          {sortedRootIssues.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-gray-500 dark:text-gray-400 mb-4">이슈가 없습니다.</p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="text-blue-600 dark:text-blue-400 hover:underline"
              >
                첫 번째 이슈 만들기
              </button>
            </div>
          ) : (
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {sortedRootIssues.map((issue) => (
                <IssueTreeItem key={issue.id} issue={issue} projectId={projectId} />
              ))}
            </div>
          )}
        </div>

        {/* Issue Count */}
        <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
          총 {issues.length}개의 이슈
        </div>
      </div>

      <CreateIssueModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        projectId={projectId}
        onSuccess={() => {
          setShowCreateModal(false);
        }}
      />
    </>
  );
}
