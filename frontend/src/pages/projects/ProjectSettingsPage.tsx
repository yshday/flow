import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useProject } from '../../hooks/useProjects';
import { useLabels, useDeleteLabel } from '../../hooks/useIssues';
import { useMilestones, useDeleteMilestone } from '../../hooks/useMilestones';
import { useProjectMembers, useRemoveMember, useUpdateMemberRole, useAddMember } from '../../hooks/useProjectMembers';
import { toast } from '../../stores/toastStore';
import { useAuthStore } from '../../stores/authStore';
import LabelModal from '../../components/label/LabelModal';
import MilestoneModal from '../../components/milestone/MilestoneModal';
import LoadingSpinner from '../../components/common/LoadingSpinner';
import ErrorState from '../../components/common/ErrorState';
import UserSection from '../../components/common/UserSection';
import type { Label, Milestone, ProjectMember, ProjectRole } from '../../types';

export default function ProjectSettingsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);

  const projectId = parseInt(id || '0');

  const { data: project, isLoading: projectLoading } = useProject(projectId);
  const { data: labels, isLoading: labelsLoading } = useLabels(projectId);
  const { data: milestones, isLoading: milestonesLoading } = useMilestones(projectId);
  const { data: members, isLoading: membersLoading } = useProjectMembers(projectId);
  const { mutateAsync: deleteLabel } = useDeleteLabel(projectId);
  const { mutateAsync: deleteMilestone } = useDeleteMilestone(projectId);
  const { mutateAsync: removeMember } = useRemoveMember(projectId);
  const { mutateAsync: updateMemberRole } = useUpdateMemberRole(projectId);
  const { mutateAsync: addMember } = useAddMember(projectId);

  const [activeTab, setActiveTab] = useState<'labels' | 'milestones' | 'members'>('labels');
  const [showAddMemberModal, setShowAddMemberModal] = useState(false);
  const [newMemberUserId, setNewMemberUserId] = useState('');
  const [newMemberRole, setNewMemberRole] = useState<ProjectRole>('member');
  const [isLabelModalOpen, setIsLabelModalOpen] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState<Label | null>(null);
  const [isMilestoneModalOpen, setIsMilestoneModalOpen] = useState(false);
  const [selectedMilestone, setSelectedMilestone] = useState<Milestone | null>(null);

  const handleEditLabel = (label: Label) => {
    setSelectedLabel(label);
    setIsLabelModalOpen(true);
  };

  const handleDeleteLabel = async (labelId: number) => {
    if (!window.confirm('이 라벨을 삭제하시겠습니까?')) return;

    try {
      await deleteLabel(labelId);
      toast.success('라벨이 삭제되었습니다.');
    } catch (error) {
      console.error('Failed to delete label:', error);
      toast.error('라벨 삭제에 실패했습니다.');
    }
  };

  const handleEditMilestone = (milestone: Milestone) => {
    setSelectedMilestone(milestone);
    setIsMilestoneModalOpen(true);
  };

  const handleDeleteMilestone = async (milestoneId: number) => {
    if (!window.confirm('이 마일스톤을 삭제하시겠습니까? 연결된 이슈는 유지됩니다.')) return;

    try {
      await deleteMilestone(milestoneId);
      toast.success('마일스톤이 삭제되었습니다.');
    } catch (error) {
      console.error('Failed to delete milestone:', error);
      toast.error('마일스톤 삭제에 실패했습니다.');
    }
  };

  const handleCloseLabelModal = () => {
    setIsLabelModalOpen(false);
    setSelectedLabel(null);
  };

  const handleCloseMilestoneModal = () => {
    setIsMilestoneModalOpen(false);
    setSelectedMilestone(null);
  };

  const handleRemoveMember = async (userId: number) => {
    if (!window.confirm('이 멤버를 제거하시겠습니까?')) return;

    try {
      await removeMember(userId);
      toast.success('멤버가 제거되었습니다.');
    } catch (error) {
      console.error('Failed to remove member:', error);
      toast.error('멤버 제거에 실패했습니다.');
    }
  };

  const handleRoleChange = async (userId: number, newRole: ProjectRole) => {
    try {
      await updateMemberRole({ userId, role: newRole });
      toast.success('역할이 변경되었습니다.');
    } catch (error) {
      console.error('Failed to update role:', error);
      toast.error('역할 변경에 실패했습니다.');
    }
  };

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();

    const userId = parseInt(newMemberUserId);
    if (isNaN(userId) || userId <= 0) {
      toast.error('유효한 사용자 ID를 입력해주세요.');
      return;
    }

    // Check if member already exists
    if (members?.some((m) => m.user_id === userId)) {
      toast.error('이미 프로젝트 멤버입니다.');
      return;
    }

    try {
      await addMember({ user_id: userId, role: newMemberRole });
      toast.success('멤버가 추가되었습니다.');
      setShowAddMemberModal(false);
      setNewMemberUserId('');
      setNewMemberRole('member');
    } catch (error) {
      console.error('Failed to add member:', error);
      toast.error('멤버 추가에 실패했습니다.');
    }
  };

  const getRoleLabel = (role: ProjectRole): string => {
    const roleLabels: Record<ProjectRole, string> = {
      owner: '소유자',
      admin: '관리자',
      member: '멤버',
      viewer: '뷰어',
    };
    return roleLabels[role];
  };

  if (projectLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <ErrorState
          message="프로젝트를 찾을 수 없습니다."
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
              onClick={() => navigate(`/projects/${id}`)}
              className="text-gray-600 hover:text-gray-900"
            >
              ← {project.name}
            </button>
            <UserSection showLogout={false} />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-8">프로젝트 설정</h1>

        {/* Tabs */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('labels')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'labels'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              라벨
            </button>
            <button
              onClick={() => setActiveTab('milestones')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'milestones'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              마일스톤
            </button>
            <button
              onClick={() => setActiveTab('members')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'members'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              멤버
            </button>
          </nav>
        </div>

        {/* Labels Section */}
        {activeTab === 'labels' && (
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900">라벨</h2>
            <button
              onClick={() => setIsLabelModalOpen(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              새 라벨
            </button>
          </div>

          {labelsLoading ? (
            <LoadingSpinner className="py-4" />
          ) : labels && labels.length > 0 ? (
            <div className="space-y-3">
              {labels.map((label) => (
                <div
                  key={label.id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-md hover:bg-gray-50"
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
                      <span className="text-sm text-gray-600">{label.description}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleEditLabel(label)}
                      className="text-gray-600 hover:text-blue-600"
                      title="수정"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDeleteLabel(label.id)}
                      className="text-gray-600 hover:text-red-600"
                      title="삭제"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              라벨이 없습니다. 새 라벨을 만들어보세요.
            </div>
          )}
        </div>
        )}

        {/* Milestones Section */}
        {activeTab === 'milestones' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900">마일스톤</h2>
            <button
              onClick={() => setIsMilestoneModalOpen(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              새 마일스톤
            </button>
          </div>

          {milestonesLoading ? (
            <LoadingSpinner className="py-4" />
          ) : milestones && milestones.length > 0 ? (
            <div className="space-y-3">
              {milestones.map((milestone) => (
                <div
                  key={milestone.id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-md hover:bg-gray-50"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-medium text-gray-900">{milestone.title}</h3>
                      <span
                        className={`px-2 py-0.5 text-xs font-medium rounded ${
                          milestone.status === 'open'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {milestone.status === 'open' ? '진행 중' : '완료'}
                      </span>
                    </div>
                    {milestone.description && (
                      <p className="text-sm text-gray-600 mb-1">{milestone.description}</p>
                    )}
                    {milestone.due_date && (
                      <p className="text-xs text-gray-500">
                        마감일: {new Date(milestone.due_date).toLocaleDateString('ko-KR')}
                      </p>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleEditMilestone(milestone)}
                      className="text-gray-600 hover:text-blue-600"
                      title="수정"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDeleteMilestone(milestone.id)}
                      className="text-gray-600 hover:text-red-600"
                      title="삭제"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              마일스톤이 없습니다. 새 마일스톤을 만들어보세요.
            </div>
          )}
        </div>
        )}

        {/* Members Section */}
        {activeTab === 'members' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900">프로젝트 멤버</h2>
            <button
              onClick={() => setShowAddMemberModal(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              멤버 추가
            </button>
          </div>

          {membersLoading ? (
            <LoadingSpinner className="py-4" />
          ) : members && members.length > 0 ? (
            <div className="space-y-3">
              {members.map((member) => (
                <div
                  key={member.user_id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-md hover:bg-gray-50"
                >
                  <div className="flex items-center gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900">
                          {member.user?.username || member.user?.email}
                        </span>
                        {member.user_id === user?.id && (
                          <span className="text-xs text-gray-500">(나)</span>
                        )}
                      </div>
                      {member.user?.email && member.user?.username && (
                        <span className="text-sm text-gray-600">{member.user.email}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <select
                      value={member.role}
                      onChange={(e) => handleRoleChange(member.user_id, e.target.value as ProjectRole)}
                      disabled={member.role === 'owner' || member.user_id === user?.id}
                      className="px-3 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                    >
                      <option value="owner">소유자</option>
                      <option value="admin">관리자</option>
                      <option value="member">멤버</option>
                      <option value="viewer">뷰어</option>
                    </select>
                    {member.role !== 'owner' && member.user_id !== user?.id && (
                      <button
                        onClick={() => handleRemoveMember(member.user_id)}
                        className="text-gray-600 hover:text-red-600"
                        title="제거"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              프로젝트 멤버가 없습니다.
            </div>
          )}
        </div>
        )}
      </main>

      {/* Modals */}
      <LabelModal
        isOpen={isLabelModalOpen}
        onClose={handleCloseLabelModal}
        projectId={projectId}
        label={selectedLabel}
      />
      <MilestoneModal
        isOpen={isMilestoneModalOpen}
        onClose={handleCloseMilestoneModal}
        projectId={projectId}
        milestone={selectedMilestone}
      />

      {/* Add Member Modal */}
      {showAddMemberModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold text-gray-900">멤버 추가</h2>
              <button
                onClick={() => {
                  setShowAddMemberModal(false);
                  setNewMemberUserId('');
                  setNewMemberRole('member');
                }}
                className="text-gray-400 hover:text-gray-600"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <form onSubmit={handleAddMember}>
              <div className="space-y-4">
                <div>
                  <label htmlFor="userId" className="block text-sm font-medium text-gray-700 mb-1">
                    사용자 ID
                  </label>
                  <input
                    type="number"
                    id="userId"
                    value={newMemberUserId}
                    onChange={(e) => setNewMemberUserId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="사용자 ID를 입력하세요"
                    required
                    min="1"
                  />
                  <p className="mt-1 text-xs text-gray-500">
                    * 추가할 사용자의 ID를 입력하세요. (향후 이메일 검색 기능이 추가될 예정입니다)
                  </p>
                </div>

                <div>
                  <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1">
                    역할
                  </label>
                  <select
                    id="role"
                    value={newMemberRole}
                    onChange={(e) => setNewMemberRole(e.target.value as ProjectRole)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="admin">관리자</option>
                    <option value="member">멤버</option>
                    <option value="viewer">뷰어</option>
                  </select>
                </div>
              </div>

              <div className="mt-6 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddMemberModal(false);
                    setNewMemberUserId('');
                    setNewMemberRole('member');
                  }}
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  취소
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  추가
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
