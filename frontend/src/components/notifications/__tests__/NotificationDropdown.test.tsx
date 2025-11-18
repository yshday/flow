import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import NotificationDropdown from '../NotificationDropdown';
import * as useNotificationsHooks from '../../../hooks/useNotifications';
import * as toastStore from '../../../stores/toastStore';
import type { Notification } from '../../../types';

// Mock the hooks
vi.mock('../../../hooks/useNotifications');
vi.mock('../../../stores/toastStore');

// Mock react-router-dom's useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockNotifications: Notification[] = [
  {
    id: 1,
    user_id: 1,
    action: 'created',
    entity_type: 'issue',
    entity_id: 10,
    message: 'New issue created: Bug in login',
    is_read: false,
    created_at: new Date(Date.now() - 1000 * 60 * 5).toISOString(), // 5 minutes ago
  },
  {
    id: 2,
    user_id: 1,
    action: 'created',
    entity_type: 'comment',
    entity_id: 20,
    message: 'New comment on your issue',
    is_read: true,
    created_at: new Date(Date.now() - 1000 * 60 * 60).toISOString(), // 1 hour ago
  },
];

describe('NotificationDropdown', () => {
  const mockMarkAsRead = vi.fn();
  const mockMarkAllAsRead = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock implementations
    vi.mocked(useNotificationsHooks.useUnreadNotificationsCount).mockReturnValue({
      data: { unread_count: 1 },
    } as any);

    vi.mocked(useNotificationsHooks.useNotifications).mockReturnValue({
      data: mockNotifications,
      isLoading: false,
      error: null,
    } as any);

    vi.mocked(useNotificationsHooks.useMarkNotificationsAsRead).mockReturnValue({
      mutateAsync: mockMarkAsRead,
    } as any);

    vi.mocked(useNotificationsHooks.useMarkAllNotificationsAsRead).mockReturnValue({
      mutateAsync: mockMarkAllAsRead,
    } as any);

    vi.mocked(toastStore.toast.success).mockImplementation(() => {});
    vi.mocked(toastStore.toast.error).mockImplementation(() => {});
  });

  const renderComponent = () => {
    return render(
      <BrowserRouter>
        <NotificationDropdown />
      </BrowserRouter>
    );
  };

  describe('Notification Bell Button', () => {
    it('should render notification bell button', () => {
      renderComponent();
      const button = screen.getByRole('button', { name: '알림' });
      expect(button).toBeInTheDocument();
    });

    it('should show unread count badge when there are unread notifications', () => {
      renderComponent();
      expect(screen.getByText('1')).toBeInTheDocument();
    });

    it('should not show badge when unread count is 0', () => {
      vi.mocked(useNotificationsHooks.useUnreadNotificationsCount).mockReturnValue({
        data: { unread_count: 0 },
      } as any);

      renderComponent();
      expect(screen.queryByText('1')).not.toBeInTheDocument();
    });

    it('should show 99+ when unread count exceeds 99', () => {
      vi.mocked(useNotificationsHooks.useUnreadNotificationsCount).mockReturnValue({
        data: { unread_count: 150 },
      } as any);

      renderComponent();
      expect(screen.getByText('99+')).toBeInTheDocument();
    });
  });

  describe('Dropdown Open/Close', () => {
    it('should open dropdown when bell button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.getByRole('menu', { name: '알림 목록' })).toBeInTheDocument();
      expect(screen.getByText('알림')).toBeInTheDocument();
    });

    it('should close dropdown when bell button is clicked again', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button', { name: '알림' });

      // Open
      await user.click(button);
      expect(screen.getByRole('menu')).toBeInTheDocument();

      // Close
      await user.click(button);
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    it('should close dropdown when Escape key is pressed', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);
      expect(screen.getByRole('menu')).toBeInTheDocument();

      // Press Escape
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should close dropdown when clicking outside', async () => {
      const user = userEvent.setup();
      const { container } = renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);
      expect(screen.getByRole('menu')).toBeInTheDocument();

      // Click outside
      await user.click(container);

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should close dropdown when close button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Click close button
      const closeButton = screen.getByRole('button', { name: '닫기' });
      await user.click(closeButton);

      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show skeleton loading when notifications are loading', async () => {
      vi.mocked(useNotificationsHooks.useNotifications).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      } as any);

      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Check for skeleton loaders (they have animate-pulse class)
      const skeletons = document.querySelectorAll('.animate-pulse');
      expect(skeletons.length).toBeGreaterThan(0);
    });
  });

  describe('Error State', () => {
    it('should show error message when notifications fail to load', async () => {
      vi.mocked(useNotificationsHooks.useNotifications).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Network error'),
      } as any);

      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.getByText('알림을 불러올 수 없습니다.')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '다시 시도' })).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should show empty message when there are no notifications', async () => {
      vi.mocked(useNotificationsHooks.useNotifications).mockReturnValue({
        data: [],
        isLoading: false,
        error: null,
      } as any);

      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.getByText('알림이 없습니다.')).toBeInTheDocument();
    });
  });

  describe('Notification List', () => {
    it('should display all notifications', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.getByText('New issue created: Bug in login')).toBeInTheDocument();
      expect(screen.getByText('New comment on your issue')).toBeInTheDocument();
    });

    it('should show unread indicator for unread notifications', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Unread notification should have bg-blue-50 class
      const notifications = screen.getAllByRole('menuitem');
      expect(notifications[0]).toHaveClass('bg-blue-50');
      expect(notifications[1]).not.toHaveClass('bg-blue-50');
    });

    it('should show mark all as read button when there are unread notifications', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.getByRole('button', { name: '모두 읽음' })).toBeInTheDocument();
    });

    it('should not show mark all as read button when all notifications are read', async () => {
      vi.mocked(useNotificationsHooks.useUnreadNotificationsCount).mockReturnValue({
        data: { unread_count: 0 },
      } as any);

      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      expect(screen.queryByRole('button', { name: '모두 읽음' })).not.toBeInTheDocument();
    });
  });

  describe('Mark All As Read', () => {
    it('should call markAllAsRead when button is clicked', async () => {
      mockMarkAllAsRead.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const bellButton = screen.getByRole('button', { name: '알림' });
      await user.click(bellButton);

      // Click mark all as read
      const markAllButton = screen.getByRole('button', { name: '모두 읽음' });
      await user.click(markAllButton);

      expect(mockMarkAllAsRead).toHaveBeenCalledTimes(1);
    });

    it('should show success toast when marking all as read succeeds', async () => {
      mockMarkAllAsRead.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const bellButton = screen.getByRole('button', { name: '알림' });
      await user.click(bellButton);

      // Click mark all as read
      const markAllButton = screen.getByRole('button', { name: '모두 읽음' });
      await user.click(markAllButton);

      await waitFor(() => {
        expect(toastStore.toast.success).toHaveBeenCalledWith('모든 알림을 읽음 처리했습니다.');
      });
    });

    it('should show error toast when marking all as read fails', async () => {
      mockMarkAllAsRead.mockRejectedValue(new Error('Network error'));
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const bellButton = screen.getByRole('button', { name: '알림' });
      await user.click(bellButton);

      // Click mark all as read
      const markAllButton = screen.getByRole('button', { name: '모두 읽음' });
      await user.click(markAllButton);

      await waitFor(() => {
        expect(toastStore.toast.error).toHaveBeenCalledWith('알림 읽음 처리에 실패했습니다.');
      });
    });
  });

  describe('Notification Click Navigation', () => {
    it('should mark notification as read and navigate to issue when clicking issue notification', async () => {
      mockMarkAsRead.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Click unread notification (issue type)
      const notification = screen.getByText('New issue created: Bug in login');
      await user.click(notification);

      await waitFor(() => {
        expect(mockMarkAsRead).toHaveBeenCalledWith({ notification_ids: [1] });
        expect(mockNavigate).toHaveBeenCalledWith('/issues/10');
      });
    });

    it('should navigate to issue for comment notification', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Click read notification (comment type)
      const notification = screen.getByText('New comment on your issue');
      await user.click(notification);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/issues/20');
      });
    });

    it('should not mark already read notification as read', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Click read notification
      const notification = screen.getByText('New comment on your issue');
      await user.click(notification);

      // Should not call markAsRead for already read notification
      expect(mockMarkAsRead).not.toHaveBeenCalled();
    });

    it('should close dropdown after clicking notification', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open dropdown
      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      // Click notification
      const notification = screen.getByText('New issue created: Bug in login');
      await user.click(notification);

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA attributes on bell button', () => {
      renderComponent();
      const button = screen.getByRole('button', { name: '알림' });

      expect(button).toHaveAttribute('aria-label', '알림');
      expect(button).toHaveAttribute('aria-expanded', 'false');
      expect(button).toHaveAttribute('aria-haspopup', 'true');
    });

    it('should update aria-expanded when dropdown is opened', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button', { name: '알림' });
      expect(button).toHaveAttribute('aria-expanded', 'false');

      await user.click(button);
      expect(button).toHaveAttribute('aria-expanded', 'true');
    });

    it('should have proper role and label on dropdown menu', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      const menu = screen.getByRole('menu', { name: '알림 목록' });
      expect(menu).toBeInTheDocument();
    });

    it('should have menuitem role on each notification', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button', { name: '알림' });
      await user.click(button);

      const menuitems = screen.getAllByRole('menuitem');
      expect(menuitems).toHaveLength(2);
    });
  });
});
