import { useState, useMemo, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useProject, useBoardColumns } from '../../hooks/useProjects';
import { useIssues, useInfiniteIssues, useMoveIssue } from '../../hooks/useIssues';
import { useLabels } from '../../hooks/useIssues';
import { useMilestones } from '../../hooks/useMilestones';
import { useDebounce } from '../../hooks/useDebounce';
import { useInfiniteScroll } from '../../hooks/useInfiniteScroll';
import { generateIssueKey, getPriorityColor, getStatusColor } from '../../lib/utils';
import CreateIssueModal from '../../components/issue/CreateIssueModal';
import KanbanBoard from '../../components/board/KanbanBoard';
import UserSection from '../../components/common/UserSection';

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const projectId = parseInt(id || '0');

  // Search and filter state
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');
  const [labelFilter, setLabelFilter] = useState<string>('');
  const [milestoneFilter, setMilestoneFilter] = useState<string>('');

  const debouncedSearch = useDebounce(searchQuery, 500);

  // Build filter params
  const filterParams = useMemo(() => {
    const params: Record<string, any> = {};
    if (debouncedSearch) params.q = debouncedSearch;
    if (statusFilter) params.status = statusFilter;
    if (priorityFilter) params.priority = priorityFilter;
    if (labelFilter) params.label_id = labelFilter;
    if (milestoneFilter) params.milestone_id = milestoneFilter;
    return params;
  }, [debouncedSearch, statusFilter, priorityFilter, labelFilter, milestoneFilter]);

  const { data: project, isLoading: projectLoading } = useProject(projectId);

  // For list view: use infinite scroll with filters
  const {
    data: issuesData,
    isLoading: issuesLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteIssues(projectId, filterParams);

  // For board view: load all issues without filters
  const { data: boardIssues, isLoading: boardIssuesLoading } = useIssues(projectId);

  const { data: columns, isLoading: columnsLoading } = useBoardColumns(projectId);
  const { data: labels } = useLabels(projectId);
  const { data: milestones } = useMilestones(projectId);
  const { mutateAsync: moveIssue } = useMoveIssue();

  // Flatten all pages of issues into a single array (for list view)
  const issues = issuesData?.pages.flatMap((page) => page) ?? [];

  // Total count from all pages
  const totalCount = issues.length;

  // Setup infinite scroll
  const sentinelRef = useInfiniteScroll({
    onLoadMore: fetchNextPage,
    hasNextPage: hasNextPage ?? false,
    isLoading: isFetchingNextPage,
  });

  const [view, setView] = useState<'list' | 'board'>('list');
  const [isCreateIssueModalOpen, setIsCreateIssueModalOpen] = useState(false);

  // Check if any filters are active
  const hasActiveFilters = searchQuery || statusFilter || priorityFilter || labelFilter || milestoneFilter;

  // Clear all filters
  const handleClearFilters = useCallback(() => {
    setSearchQuery('');
    setStatusFilter('');
    setPriorityFilter('');
    setLabelFilter('');
    setMilestoneFilter('');
  }, []);

  const handleIssueMove = useCallback(
    async (issueId: number, columnId: number, version: number) => {
      try {
        // Find the target column to determine status
        const targetColumn = columns?.find((col) => col.id === columnId);

        // Auto-set status based on column name
        let status: 'open' | 'in_progress' | 'closed' = 'open';
        const columnName = targetColumn?.name.toLowerCase();

        if (columnName === 'done') {
          status = 'closed';
        } else if (columnName === 'in progress') {
          status = 'in_progress';
        } else {
          status = 'open';
        }

        await moveIssue({
          id: issueId,
          data: {
            column_id: columnId,
            version,
            position: 0,
            status,
          },
        });
      } catch (error) {
        console.error('Failed to move issue:', error);
      }
    },
    [moveIssue, columns]
  );

  if (projectLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">로딩 중...</div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg text-red-600">프로젝트를 찾을 수 없습니다.</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/projects')}
                className="text-gray-600 hover:text-gray-900"
              >
                ← 프로젝트 목록
              </button>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{project.name}</h1>
                <p className="text-sm text-gray-500">{project.key}</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <Link
                to={`/projects/${id}/settings`}
                className="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                  />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                </svg>
                설정
              </Link>
              <UserSection />
            </div>
          </div>
        </div>
      </header>

      {/* Sub Navigation */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8">
            <button
              onClick={() => setView('list')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                view === 'list'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              이슈 목록
            </button>
            <button
              onClick={() => setView('board')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                view === 'board'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              칸반 보드
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {view === 'list' ? (
          <>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-semibold text-gray-900">이슈</h2>
              <button
                onClick={() => setIsCreateIssueModalOpen(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                새 이슈
              </button>
            </div>

            {/* Search and Filters */}
            <div className="bg-white shadow rounded-lg p-4 mb-6">
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                {/* Search Input */}
                <div className="md:col-span-2">
                  <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
                    검색
                  </label>
                  <input
                    id="search"
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="제목 또는 설명 검색..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                {/* Status Filter */}
                <div>
                  <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
                    상태
                  </label>
                  <select
                    id="status"
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">전체</option>
                    <option value="open">열림</option>
                    <option value="in_progress">진행 중</option>
                    <option value="closed">닫힘</option>
                  </select>
                </div>

                {/* Priority Filter */}
                <div>
                  <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-1">
                    우선순위
                  </label>
                  <select
                    id="priority"
                    value={priorityFilter}
                    onChange={(e) => setPriorityFilter(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">전체</option>
                    <option value="low">낮음</option>
                    <option value="medium">보통</option>
                    <option value="high">높음</option>
                    <option value="urgent">긴급</option>
                  </select>
                </div>

                {/* Label Filter */}
                <div>
                  <label htmlFor="label" className="block text-sm font-medium text-gray-700 mb-1">
                    라벨
                  </label>
                  <select
                    id="label"
                    value={labelFilter}
                    onChange={(e) => setLabelFilter(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">전체</option>
                    {labels?.map((label) => (
                      <option key={label.id} value={label.id}>
                        {label.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Milestone Filter (Second Row) */}
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mt-4">
                <div className="md:col-span-2">
                  <label htmlFor="milestone" className="block text-sm font-medium text-gray-700 mb-1">
                    마일스톤
                  </label>
                  <select
                    id="milestone"
                    value={milestoneFilter}
                    onChange={(e) => setMilestoneFilter(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">전체</option>
                    {milestones?.map((milestone) => (
                      <option key={milestone.id} value={milestone.id}>
                        {milestone.title}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Results Count and Clear Button */}
                <div className="md:col-span-3 flex items-end justify-between">
                  <div className="text-sm text-gray-600">
                    {totalCount > 0 ? `${totalCount}개의 이슈${hasNextPage ? '+' : ''}` : issuesLoading ? '로딩 중...' : '0개의 이슈'}
                  </div>
                  {hasActiveFilters && (
                    <button
                      onClick={handleClearFilters}
                      className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                    >
                      필터 초기화
                    </button>
                  )}
                </div>
              </div>
            </div>

            {issuesLoading ? (
              <div className="text-center py-12">로딩 중...</div>
            ) : issues && issues.length > 0 ? (
              <>
                <div className="bg-white shadow rounded-lg overflow-hidden">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          이슈
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          제목
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          상태
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          우선순위
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          생성일
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {issues.map((issue) => (
                        <tr
                          key={issue.id}
                          className="hover:bg-gray-50 cursor-pointer"
                          onClick={() => navigate(`/projects/${projectId}/issues/${issue.id}`)}
                        >
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
                            {generateIssueKey(project.key, issue.issue_number)}
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-900">
                            {issue.title}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor(
                                issue.status
                              )}`}
                            >
                              {issue.status === 'open' ? '열림' : issue.status === 'in_progress' ? '진행 중' : '닫힘'}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`px-2 py-1 text-xs font-medium rounded ${getPriorityColor(
                                issue.priority
                              )}`}
                            >
                              {issue.priority}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {new Date(issue.created_at).toLocaleDateString('ko-KR')}
                          </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Sentinel element for infinite scroll */}
              <div ref={sentinelRef} className="h-10" />

              {/* Loading indicator */}
              {isFetchingNextPage && (
                <div className="py-4 text-center text-sm text-gray-600">
                  더 많은 이슈를 불러오는 중...
                </div>
              )}

              {/* End of list indicator */}
              {!hasNextPage && totalCount > 0 && (
                <div className="py-4 text-center text-sm text-gray-500">
                  모든 이슈를 불러왔습니다
                </div>
              )}
            </>
            ) : (
              <div className="bg-white p-12 rounded-lg shadow text-center">
                <p className="text-gray-600">이슈가 없습니다.</p>
                <button
                  onClick={() => setIsCreateIssueModalOpen(true)}
                  className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  첫 번째 이슈 만들기
                </button>
              </div>
            )}
          </>
        ) : (
          <>
            {columnsLoading || boardIssuesLoading ? (
              <div className="text-center py-12">로딩 중...</div>
            ) : columns && boardIssues ? (
              <KanbanBoard
                columns={columns}
                issues={boardIssues}
                projectKey={project.key}
                projectId={projectId}
                onIssueMove={handleIssueMove}
                onIssueClick={(issueId) =>
                  navigate(`/projects/${projectId}/issues/${issueId}`)
                }
              />
            ) : (
              <div className="bg-white p-12 rounded-lg shadow text-center">
                <p className="text-gray-600">데이터를 불러올 수 없습니다.</p>
              </div>
            )}
          </>
        )}
      </main>

      {/* Create Issue Modal */}
      <CreateIssueModal
        isOpen={isCreateIssueModalOpen}
        onClose={() => setIsCreateIssueModalOpen(false)}
        projectId={projectId}
      />
    </div>
  );
}
