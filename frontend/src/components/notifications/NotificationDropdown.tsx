import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useNotifications, useUnreadNotificationsCount, useMarkNotificationsAsRead, useMarkAllNotificationsAsRead } from '../../hooks/useNotifications';
import { toast } from '../../stores/toastStore';
import { formatRelativeTime } from '../../lib/utils';
import type { Notification } from '../../types';

export default function NotificationDropdown() {
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const { data: unreadCount } = useUnreadNotificationsCount();
  const { data: notifications, isLoading, error } = useNotifications({ limit: 20 });
  const { mutateAsync: markAsRead } = useMarkNotificationsAsRead();
  const { mutateAsync: markAllAsRead } = useMarkAllNotificationsAsRead();

  // Close dropdown when clicking outside or pressing Escape
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleEscapeKey(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscapeKey);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscapeKey);
    };
  }, [isOpen]);

  const handleNotificationClick = async (notification: Notification) => {
    // Mark as read if not already read
    if (!notification.is_read) {
      try {
        await markAsRead({ notification_ids: [notification.id] });
      } catch (error) {
        console.error('Failed to mark notification as read:', error);
      }
    }

    // Navigate to the related entity
    // Close the dropdown
    setIsOpen(false);

    // Navigate based on entity type
    switch (notification.entity_type) {
      case 'issue':
        navigate(`/issues/${notification.entity_id}`);
        break;
      case 'comment':
        // Comments belong to issues, so navigate to the issue
        // We would need issue_id from the notification, but for now just use entity_id
        navigate(`/issues/${notification.entity_id}`);
        break;
      case 'project':
        navigate(`/projects/${notification.entity_id}`);
        break;
      case 'label':
      case 'milestone':
      case 'member':
        // Navigate to project settings
        // We would need project_id, but for now we can't navigate accurately
        // Just mark as read for now
        break;
      default:
        break;
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead();
      toast.success('모든 알림을 읽음 처리했습니다.');
    } catch (error) {
      console.error('Failed to mark all as read:', error);
      toast.error('알림 읽음 처리에 실패했습니다.');
    }
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell Icon Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-gray-600 hover:text-gray-900 focus:outline-none"
        aria-label="알림"
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        <svg
          className="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
          />
        </svg>

        {/* Unread Badge */}
        {unreadCount && unreadCount.unread_count > 0 && (
          <span className="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-600 rounded-full">
            {unreadCount.unread_count > 99 ? '99+' : unreadCount.unread_count}
          </span>
        )}
      </button>

      {/* Dropdown Panel */}
      {isOpen && (
        <div
          className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-lg border border-gray-200 z-50"
          role="menu"
          aria-label="알림 목록"
        >
          {/* Header */}
          <div className="px-4 py-3 border-b border-gray-200 flex justify-between items-center">
            <h3 className="text-lg font-semibold text-gray-900">알림</h3>
            {unreadCount && unreadCount.unread_count > 0 && (
              <button
                onClick={handleMarkAllAsRead}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                모두 읽음
              </button>
            )}
          </div>

          {/* Notifications List */}
          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <div className="divide-y divide-gray-200">
                {/* Skeleton Loading */}
                {[1, 2, 3].map((i) => (
                  <div key={i} className="px-4 py-3">
                    <div className="flex items-start gap-3">
                      <div className="w-2 h-2 bg-gray-200 rounded-full mt-2 flex-shrink-0 animate-pulse" />
                      <div className="flex-1 min-w-0 space-y-2">
                        <div className="h-4 bg-gray-200 rounded animate-pulse w-3/4" />
                        <div className="h-3 bg-gray-200 rounded animate-pulse w-1/4" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : error ? (
              <div className="px-4 py-8 text-center">
                <p className="text-sm text-red-600 mb-2">알림을 불러올 수 없습니다.</p>
                <button
                  onClick={() => window.location.reload()}
                  className="text-xs text-blue-600 hover:text-blue-800"
                >
                  다시 시도
                </button>
              </div>
            ) : notifications && notifications.length > 0 ? (
              <div className="divide-y divide-gray-200">
                {notifications.map((notification) => (
                  <div
                    key={notification.id}
                    onClick={() => handleNotificationClick(notification)}
                    className={`px-4 py-3 hover:bg-gray-50 cursor-pointer ${
                      !notification.is_read ? 'bg-blue-50' : ''
                    }`}
                    role="menuitem"
                  >
                    <div className="flex items-start gap-3">
                      {!notification.is_read && (
                        <div className="w-2 h-2 bg-blue-600 rounded-full mt-2 flex-shrink-0" />
                      )}
                      <div className="flex-1 min-w-0">
                        <p className={`text-sm ${!notification.is_read ? 'font-medium text-gray-900' : 'text-gray-700'}`}>
                          {notification.message}
                        </p>
                        <p className="text-xs text-gray-500 mt-1">
                          {formatRelativeTime(notification.created_at)}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="px-4 py-8 text-center text-gray-500">
                알림이 없습니다.
              </div>
            )}
          </div>

          {/* Footer */}
          {notifications && notifications.length > 0 && (
            <div className="px-4 py-3 border-t border-gray-200 text-center">
              <button
                onClick={() => setIsOpen(false)}
                className="text-sm text-gray-600 hover:text-gray-900"
              >
                닫기
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
