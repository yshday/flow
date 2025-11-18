import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useIssue, useUpdateIssue, useComments, useCreateComment, useIssueLabels, useLabels, useAddLabelToIssue, useRemoveLabelFromIssue } from '../../hooks/useIssues';
import { useProject } from '../../hooks/useProjects';
import { useMilestone, useMilestones } from '../../hooks/useMilestones';
import { useAuthStore } from '../../stores/authStore';
import { toast } from '../../stores/toastStore';
import { generateIssueKey, getPriorityColor, getStatusColor } from '../../lib/utils';
import type { UpdateIssueRequest } from '../../types';
import ActivityLog from '../../components/issue/ActivityLog';
import LoadingSpinner from '../../components/common/LoadingSpinner';
import ErrorState from '../../components/common/ErrorState';

export default function IssueDetailPage() {
  const { projectId, issueId } = useParams<{ projectId: string; issueId: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);

  const parsedProjectId = parseInt(projectId || '0');
  const parsedIssueId = parseInt(issueId || '0');

  const { data: project } = useProject(parsedProjectId);
  const { data: issue, isLoading: issueLoading } = useIssue(parsedIssueId);
  const { data: comments, isLoading: commentsLoading } = useComments(parsedIssueId);
  const { data: issueLabels } = useIssueLabels(parsedIssueId);
  const { data: projectLabels } = useLabels(parsedProjectId);
  const { data: milestone } = useMilestone(issue?.milestone_id || 0);
  const { data: projectMilestones } = useMilestones(parsedProjectId);
  const { mutateAsync: updateIssue } = useUpdateIssue(parsedIssueId);
  const { mutateAsync: createComment } = useCreateComment(parsedIssueId);
  const { mutateAsync: addLabel } = useAddLabelToIssue(parsedIssueId);
  const { mutateAsync: removeLabel } = useRemoveLabelFromIssue(parsedIssueId);

  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    title: '',
    description: '',
    status: 'open' as 'open' | 'closed',
    priority: 'medium' as 'low' | 'medium' | 'high' | 'urgent',
  });

  const [newComment, setNewComment] = useState('');

  // Initialize edit form when issue loads
  if (issue && !isEditing && editForm.title === '') {
    setEditForm({
      title: issue.title,
      description: issue.description || '',
      status: issue.status,
      priority: issue.priority,
    });
  }

  const handleEdit = () => {
    if (issue) {
      setEditForm({
        title: issue.title,
        description: issue.description || '',
        status: issue.status,
        priority: issue.priority,
      });
      setIsEditing(true);
    }
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
  };

  const handleSaveEdit = async () => {
    if (!issue) return;

    try {
      const updateData: UpdateIssueRequest = {
        title: editForm.title !== issue.title ? editForm.title : undefined,
        description: editForm.description !== issue.description ? editForm.description : undefined,
        status: editForm.status !== issue.status ? editForm.status : undefined,
        priority: editForm.priority !== issue.priority ? editForm.priority : undefined,
      };

      await updateIssue(updateData);
      setIsEditing(false);
      toast.success('이슈가 성공적으로 수정되었습니다.');
    } catch (error) {
      console.error('Failed to update issue:', error);
      toast.error('이슈 수정에 실패했습니다.');
    }
  };

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    try {
      await createComment(newComment);
      setNewComment('');
      toast.success('댓글이 작성되었습니다.');
    } catch (error) {
      console.error('Failed to create comment:', error);
      toast.error('댓글 작성에 실패했습니다.');
    }
  };

  const handleAddLabel = async (labelId: number) => {
    try {
      await addLabel(labelId);
      toast.success('라벨이 추가되었습니다.');
    } catch (error) {
      console.error('Failed to add label:', error);
      toast.error('라벨 추가에 실패했습니다.');
    }
  };

  const handleRemoveLabel = async (labelId: number) => {
    try {
      await removeLabel(labelId);
      toast.success('라벨이 제거되었습니다.');
    } catch (error) {
      console.error('Failed to remove label:', error);
      toast.error('라벨 제거에 실패했습니다.');
    }
  };

  const handleChangeMilestone = async (milestoneId: number | null) => {
    try {
      await updateIssue({ milestone_id: milestoneId || undefined });
      toast.success(milestoneId ? '마일스톤이 할당되었습니다.' : '마일스톤이 제거되었습니다.');
    } catch (error) {
      console.error('Failed to change milestone:', error);
      toast.error('마일스톤 변경에 실패했습니다.');
    }
  };

  // Get labels that are not yet assigned to this issue
  const availableLabels = projectLabels?.filter(
    (label) => !issueLabels?.some((issueLabel) => issueLabel.id === label.id)
  ) || [];

  if (issueLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!issue || !project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <ErrorState
          message="이슈를 찾을 수 없습니다."
          onRetry={() => navigate('/projects')}
        />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <button
              onClick={() => navigate(`/projects/${projectId}`)}
              className="text-gray-600 hover:text-gray-900"
            >
              ← {project.name}
            </button>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">{user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Issue Header */}
            <div className="bg-white rounded-lg shadow p-6">
              <div className="flex items-center space-x-2 mb-4">
                <span className="text-sm font-medium text-blue-600">
                  {generateIssueKey(project.key, issue.issue_number)}
                </span>
                <span
                  className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor(issue.status)}`}
                >
                  {issue.status === 'open' ? '열림' : '닫힘'}
                </span>
              </div>

              {isEditing ? (
                <div className="space-y-4">
                  <input
                    type="text"
                    value={editForm.title}
                    onChange={(e) => setEditForm({ ...editForm, title: e.target.value })}
                    className="w-full text-2xl font-bold border-b border-gray-300 focus:outline-none focus:border-blue-500 pb-2"
                  />
                  <textarea
                    value={editForm.description}
                    onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                    rows={6}
                    className="w-full border border-gray-300 rounded-md p-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="이슈 설명을 입력하세요..."
                  />
                  <div className="flex space-x-4">
                    <button
                      onClick={handleSaveEdit}
                      className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                    >
                      저장
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
                    >
                      취소
                    </button>
                  </div>
                </div>
              ) : (
                <div>
                  <div className="flex justify-between items-start mb-4">
                    <h1 className="text-2xl font-bold text-gray-900">{issue.title}</h1>
                    <button
                      onClick={handleEdit}
                      className="text-sm text-blue-600 hover:text-blue-700"
                    >
                      수정
                    </button>
                  </div>
                  <p className="text-gray-700 whitespace-pre-wrap">
                    {issue.description || '설명이 없습니다.'}
                  </p>
                </div>
              )}
            </div>

            {/* Comments Section */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">댓글</h2>

              {/* Comment Form */}
              <form onSubmit={handleSubmitComment} className="mb-6">
                <textarea
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  rows={3}
                  className="w-full border border-gray-300 rounded-md p-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="댓글을 입력하세요..."
                />
                <div className="mt-2 flex justify-end">
                  <button
                    type="submit"
                    disabled={!newComment.trim()}
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    댓글 작성
                  </button>
                </div>
              </form>

              {/* Comment List */}
              {commentsLoading ? (
                <LoadingSpinner className="py-4" />
              ) : comments && comments.length > 0 ? (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border-t border-gray-200 pt-4">
                      <div className="flex justify-between items-start mb-2">
                        <span className="text-sm font-medium text-gray-900">
                          사용자 #{comment.user_id}
                        </span>
                        <span className="text-sm text-gray-500">
                          {new Date(comment.created_at).toLocaleString('ko-KR')}
                        </span>
                      </div>
                      <p className="text-gray-700 whitespace-pre-wrap">{comment.content}</p>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  아직 댓글이 없습니다.
                </div>
              )}
            </div>

            {/* Activity Log Section */}
            <div className="bg-white rounded-lg shadow p-6">
              <ActivityLog issueId={parsedIssueId} />
            </div>
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow p-6 space-y-4">
              <h3 className="font-semibold text-gray-900">상세 정보</h3>

              {/* Priority */}
              <div>
                <label className="text-sm font-medium text-gray-700">우선순위</label>
                {isEditing ? (
                  <select
                    value={editForm.priority}
                    onChange={(e) => setEditForm({ ...editForm, priority: e.target.value as any })}
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="low">낮음 (Low)</option>
                    <option value="medium">보통 (Medium)</option>
                    <option value="high">높음 (High)</option>
                    <option value="urgent">긴급 (Urgent)</option>
                  </select>
                ) : (
                  <div className="mt-1">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${getPriorityColor(issue.priority)}`}>
                      {issue.priority}
                    </span>
                  </div>
                )}
              </div>

              {/* Status */}
              <div>
                <label className="text-sm font-medium text-gray-700">상태</label>
                {isEditing ? (
                  <select
                    value={editForm.status}
                    onChange={(e) => setEditForm({ ...editForm, status: e.target.value as any })}
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="open">열림</option>
                    <option value="closed">닫힘</option>
                  </select>
                ) : (
                  <div className="mt-1">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor(issue.status)}`}>
                      {issue.status === 'open' ? '열림' : '닫힘'}
                    </span>
                  </div>
                )}
              </div>

              {/* Labels */}
              <div>
                <label className="text-sm font-medium text-gray-700">라벨</label>
                <div className="mt-1 space-y-2">
                  {/* Current Labels */}
                  <div className="flex flex-wrap gap-1">
                    {issueLabels && issueLabels.length > 0 ? (
                      issueLabels.map((label) => (
                        <span
                          key={label.id}
                          className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded"
                          style={{
                            backgroundColor: label.color + '20',
                            color: label.color,
                            border: `1px solid ${label.color}`,
                          }}
                        >
                          {label.name}
                          <button
                            onClick={() => handleRemoveLabel(label.id)}
                            className="hover:opacity-70"
                            title="라벨 제거"
                          >
                            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </button>
                        </span>
                      ))
                    ) : (
                      <p className="text-sm text-gray-500">라벨 없음</p>
                    )}
                  </div>

                  {/* Add Label Dropdown */}
                  {availableLabels.length > 0 && (
                    <select
                      onChange={(e) => {
                        const labelId = parseInt(e.target.value);
                        if (labelId) {
                          handleAddLabel(labelId);
                          e.target.value = '';
                        }
                      }}
                      className="w-full text-sm border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      defaultValue=""
                    >
                      <option value="" disabled>라벨 추가...</option>
                      {availableLabels.map((label) => (
                        <option key={label.id} value={label.id}>
                          {label.name}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              </div>

              {/* Milestone */}
              <div>
                <label className="text-sm font-medium text-gray-700">마일스톤</label>
                <div className="mt-1 space-y-2">
                  {/* Current Milestone */}
                  {milestone ? (
                    <div className="p-2 bg-gray-50 rounded border border-gray-200">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <p className="text-sm font-medium text-gray-900">{milestone.title}</p>
                          {milestone.due_date && (
                            <p className="text-xs text-gray-500 mt-1">
                              {new Date(milestone.due_date).toLocaleDateString('ko-KR')}
                            </p>
                          )}
                        </div>
                        <button
                          onClick={() => handleChangeMilestone(null)}
                          className="text-gray-400 hover:text-gray-600"
                          title="마일스톤 제거"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  ) : (
                    <p className="text-sm text-gray-500">마일스톤 없음</p>
                  )}

                  {/* Change Milestone Dropdown */}
                  {projectMilestones && projectMilestones.length > 0 && (
                    <select
                      onChange={(e) => {
                        const milestoneId = e.target.value ? parseInt(e.target.value) : null;
                        if (milestoneId !== issue?.milestone_id) {
                          handleChangeMilestone(milestoneId);
                          e.target.value = issue?.milestone_id?.toString() || '';
                        }
                      }}
                      value={issue?.milestone_id || ''}
                      className="w-full text-sm border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="">마일스톤 선택...</option>
                      {projectMilestones.map((ms) => (
                        <option key={ms.id} value={ms.id}>
                          {ms.title}
                          {ms.due_date && ` (${new Date(ms.due_date).toLocaleDateString('ko-KR')})`}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              </div>

              {/* Reporter */}
              <div>
                <label className="text-sm font-medium text-gray-700">보고자</label>
                <p className="mt-1 text-sm text-gray-600">사용자 #{issue.reporter_id}</p>
              </div>

              {/* Dates */}
              <div>
                <label className="text-sm font-medium text-gray-700">생성일</label>
                <p className="mt-1 text-sm text-gray-600">
                  {new Date(issue.created_at).toLocaleString('ko-KR')}
                </p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-700">수정일</label>
                <p className="mt-1 text-sm text-gray-600">
                  {new Date(issue.updated_at).toLocaleString('ko-KR')}
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
