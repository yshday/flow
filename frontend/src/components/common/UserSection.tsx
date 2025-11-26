import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../stores/authStore';
import { useAuth } from '../../hooks/useAuth';
import NotificationDropdown from '../notifications/NotificationDropdown';

interface UserSectionProps {
  showLogout?: boolean;
}

export default function UserSection({ showLogout = true }: UserSectionProps) {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const { logout } = useAuth();

  return (
    <div className="flex items-center space-x-4">
      <NotificationDropdown />
      <button
        onClick={() => navigate('/profile')}
        className="text-sm text-gray-600 hover:text-gray-900"
      >
        {user?.username || user?.email}
      </button>
      {showLogout && (
        <button
          onClick={logout}
          className="text-sm text-gray-600 hover:text-gray-900"
        >
          로그아웃
        </button>
      )}
    </div>
  );
}
