import { describe, it, expect, beforeEach, vi } from 'vitest';
import { notificationsApi } from '../notifications';
import { apiClient } from '../client';
import type { Notification } from '../../types';

// Mock apiClient
vi.mock('../client', () => ({
  apiClient: {
    get: vi.fn(),
    put: vi.fn(),
  },
}));

describe('notificationsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('list', () => {
    it('should fetch notifications with default parameters', async () => {
      const mockNotifications: Notification[] = [
        {
          id: 1,
          user_id: 1,
          action: 'created',
          entity_type: 'issue',
          entity_id: 1,
          message: 'Test notification',
          is_read: false,
          created_at: '2025-01-01T00:00:00Z',
        },
      ];

      vi.mocked(apiClient.get).mockResolvedValue({ data: mockNotifications });

      const result = await notificationsApi.list();

      expect(apiClient.get).toHaveBeenCalledWith('/notifications', { params: undefined });
      expect(result).toEqual(mockNotifications);
    });

    it('should fetch notifications with filters', async () => {
      const mockNotifications: Notification[] = [];
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockNotifications });

      const params = { unread: true, limit: 10, offset: 0 };
      await notificationsApi.list(params);

      expect(apiClient.get).toHaveBeenCalledWith('/notifications', { params });
    });
  });

  describe('getUnreadCount', () => {
    it('should fetch unread notification count', async () => {
      const mockCount = { unread_count: 5 };
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockCount });

      const result = await notificationsApi.getUnreadCount();

      expect(apiClient.get).toHaveBeenCalledWith('/notifications/unread/count');
      expect(result).toEqual(mockCount);
    });
  });

  describe('markAsRead', () => {
    it('should mark notifications as read', async () => {
      vi.mocked(apiClient.put).mockResolvedValue({});

      const data = { notification_ids: [1, 2, 3] };
      await notificationsApi.markAsRead(data);

      expect(apiClient.put).toHaveBeenCalledWith('/notifications/read', data);
    });
  });

  describe('markAllAsRead', () => {
    it('should call the optimized mark all as read endpoint', async () => {
      vi.mocked(apiClient.put).mockResolvedValue({});

      await notificationsApi.markAllAsRead();

      expect(apiClient.put).toHaveBeenCalledWith('/notifications/read/all');
      expect(apiClient.put).toHaveBeenCalledTimes(1);
    });

    it('should not fetch notifications list (optimized)', async () => {
      vi.mocked(apiClient.put).mockResolvedValue({});

      await notificationsApi.markAllAsRead();

      // Verify that we don't call GET /notifications
      expect(apiClient.get).not.toHaveBeenCalled();
    });
  });
});
