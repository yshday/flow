import { useNavigate } from 'react-router-dom';
import { useCurrentUser, useUserMemberships } from '../../hooks/useAuth';
import { useAuthStore } from '../../stores/authStore';
import LoadingSpinner from '../../components/common/LoadingSpinner';
import ErrorState from '../../components/common/ErrorState';

export default function ProfilePage() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const { data: currentUser, isLoading: userLoading } = useCurrentUser();
  const { data: memberships, isLoading: membershipsLoading } = useUserMemberships();

  if (userLoading || membershipsLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!currentUser) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <ErrorState
          message="사용자 정보를 찾을 수 없습니다."
          onRetry={() => navigate('/')}
        />
      </div>
    );
  }

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'owner':
        return 'bg-purple-100 text-purple-800';
      case 'admin':
        return 'bg-blue-100 text-blue-800';
      case 'member':
        return 'bg-green-100 text-green-800';
      case 'viewer':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getRoleText = (role: string) => {
    switch (role) {
      case 'owner':
        return '소유자';
      case 'admin':
        return '관리자';
      case 'member':
        return '멤버';
      case 'viewer':
        return '뷰어';
      default:
        return role;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <button
              onClick={() => navigate('/projects')}
              className="text-gray-600 hover:text-gray-900"
            >
              ← 프로젝트 목록
            </button>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">{user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-6">
          {/* User Info Card */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">프로필</h2>
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-gray-700">이메일</label>
                <p className="mt-1 text-gray-900">{currentUser.email}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">사용자명</label>
                <p className="mt-1 text-gray-900">{currentUser.username}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-700">가입일</label>
                <p className="mt-1 text-gray-900">
                  {new Date(currentUser.created_at).toLocaleString('ko-KR')}
                </p>
              </div>
            </div>
          </div>

          {/* Memberships Card */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-4">프로젝트 권한</h2>
            {memberships && memberships.length > 0 ? (
              <div className="space-y-3">
                {memberships.map((membership) => (
                  <div
                    key={membership.project_id}
                    className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/projects/${membership.project_id}`)}
                  >
                    <div className="flex-1">
                      <h3 className="font-semibold text-gray-900">{membership.project.name}</h3>
                      <p className="text-sm text-gray-600">
                        키: <span className="font-mono">{membership.project.key}</span>
                      </p>
                      {membership.project.description && (
                        <p className="text-sm text-gray-500 mt-1">
                          {membership.project.description}
                        </p>
                      )}
                    </div>
                    <div className="ml-4">
                      <span
                        className={`px-3 py-1 text-sm font-medium rounded-full ${getRoleBadgeColor(
                          membership.role
                        )}`}
                      >
                        {getRoleText(membership.role)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                아직 소속된 프로젝트가 없습니다.
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
