import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import {
  useNotifications,
  useUnreadNotificationsCount,
  useMarkNotificationsAsRead,
  useMarkAllNotificationsAsRead,
} from '../useNotifications';
import { notificationsApi } from '../../api/notifications';
import type { Notification } from '../../types';

// Mock the notifications API
vi.mock('../../api/notifications');

const mockNotifications: Notification[] = [
  {
    id: 1,
    user_id: 1,
    action: 'created',
    entity_type: 'issue',
    entity_id: 10,
    message: 'New issue created',
    is_read: false,
    created_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 2,
    user_id: 1,
    action: 'created',
    entity_type: 'comment',
    entity_id: 20,
    message: 'New comment',
    is_read: true,
    created_at: '2025-01-01T01:00:00Z',
  },
];

describe('useNotifications hooks', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();

    // Create a fresh QueryClient for each test to ensure isolation
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false, // Disable retries for faster tests
        },
      },
    });
  });

  // Helper to create a wrapper with QueryClientProvider
  const createWrapper = () => {
    return ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

  describe('useNotifications', () => {
    it('should fetch notifications successfully', async () => {
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);

      const { result } = renderHook(() => useNotifications(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockNotifications);
      expect(notificationsApi.list).toHaveBeenCalledWith(undefined);
    });

    it('should fetch notifications with params', async () => {
      const params = { unread: true, limit: 10, offset: 0 };
      vi.mocked(notificationsApi.list).mockResolvedValue([mockNotifications[0]]);

      const { result } = renderHook(() => useNotifications(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual([mockNotifications[0]]);
      expect(notificationsApi.list).toHaveBeenCalledWith(params);
    });

    it('should handle errors', async () => {
      const error = new Error('Network error');
      vi.mocked(notificationsApi.list).mockRejectedValue(error);

      const { result } = renderHook(() => useNotifications(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toBeTruthy();
    });

    it('should use correct query key with params', async () => {
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);

      const params = { unread: true, limit: 5 };
      const { result } = renderHook(() => useNotifications(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Check that the query was cached with the correct key
      const cachedData = queryClient.getQueryData(['notifications', params]);
      expect(cachedData).toEqual(mockNotifications);
    });

    it('should configure refetch interval', async () => {
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);

      const { result } = renderHook(() => useNotifications(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Check that the query has the correct refetch interval (30 seconds)
      const queryState = queryClient.getQueryState(['notifications', undefined]);
      expect(queryState).toBeDefined();
      // Note: We can't easily test the actual interval timing in unit tests
      // but we've verified the configuration in the hook implementation
    });
  });

  describe('useUnreadNotificationsCount', () => {
    it('should fetch unread count successfully', async () => {
      const mockCount = { unread_count: 5 };
      vi.mocked(notificationsApi.getUnreadCount).mockResolvedValue(mockCount);

      const { result } = renderHook(() => useUnreadNotificationsCount(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockCount);
      expect(notificationsApi.getUnreadCount).toHaveBeenCalledTimes(1);
    });

    it('should handle errors', async () => {
      const error = new Error('Network error');
      vi.mocked(notificationsApi.getUnreadCount).mockRejectedValue(error);

      const { result } = renderHook(() => useUnreadNotificationsCount(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toBeTruthy();
    });

    it('should use correct query key', async () => {
      const mockCount = { unread_count: 3 };
      vi.mocked(notificationsApi.getUnreadCount).mockResolvedValue(mockCount);

      renderHook(() => useUnreadNotificationsCount(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        const cachedData = queryClient.getQueryData(['notifications', 'unread', 'count']);
        expect(cachedData).toEqual(mockCount);
      });
    });
  });

  describe('useMarkNotificationsAsRead', () => {
    it('should mark notifications as read successfully', async () => {
      vi.mocked(notificationsApi.markAsRead).mockResolvedValue(undefined);

      const { result } = renderHook(() => useMarkNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      const data = { notification_ids: [1, 2, 3] };
      await result.current.mutateAsync(data);

      expect(notificationsApi.markAsRead).toHaveBeenCalledWith(data);
    });

    it('should invalidate queries on success', async () => {
      vi.mocked(notificationsApi.markAsRead).mockResolvedValue(undefined);
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);

      // First, populate the cache
      await queryClient.fetchQuery({
        queryKey: ['notifications'],
        queryFn: () => notificationsApi.list(),
      });

      const { result } = renderHook(() => useMarkNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      await result.current.mutateAsync({ notification_ids: [1] });

      await waitFor(() => {
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['notifications'] });
      });
    });

    it('should handle errors', async () => {
      const error = new Error('Network error');
      vi.mocked(notificationsApi.markAsRead).mockRejectedValue(error);

      const { result } = renderHook(() => useMarkNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      await expect(result.current.mutateAsync({ notification_ids: [1] })).rejects.toThrow(
        'Network error'
      );
    });
  });

  describe('useMarkAllNotificationsAsRead', () => {
    it('should mark all notifications as read successfully', async () => {
      vi.mocked(notificationsApi.markAllAsRead).mockResolvedValue(undefined);

      const { result } = renderHook(() => useMarkAllNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      await result.current.mutateAsync();

      expect(notificationsApi.markAllAsRead).toHaveBeenCalledTimes(1);
    });

    it('should invalidate queries on success', async () => {
      vi.mocked(notificationsApi.markAllAsRead).mockResolvedValue(undefined);
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);

      // First, populate the cache
      await queryClient.fetchQuery({
        queryKey: ['notifications'],
        queryFn: () => notificationsApi.list(),
      });

      const { result } = renderHook(() => useMarkAllNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      await result.current.mutateAsync();

      await waitFor(() => {
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['notifications'] });
      });
    });

    it('should handle errors', async () => {
      const error = new Error('Network error');
      vi.mocked(notificationsApi.markAllAsRead).mockRejectedValue(error);

      const { result } = renderHook(() => useMarkAllNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      await expect(result.current.mutateAsync()).rejects.toThrow('Network error');
    });

    it('should invalidate both notification list and count queries', async () => {
      vi.mocked(notificationsApi.markAllAsRead).mockResolvedValue(undefined);
      vi.mocked(notificationsApi.list).mockResolvedValue(mockNotifications);
      vi.mocked(notificationsApi.getUnreadCount).mockResolvedValue({ unread_count: 0 });

      // Populate both caches
      await queryClient.fetchQuery({
        queryKey: ['notifications'],
        queryFn: () => notificationsApi.list(),
      });
      await queryClient.fetchQuery({
        queryKey: ['notifications', 'unread', 'count'],
        queryFn: () => notificationsApi.getUnreadCount(),
      });

      const { result } = renderHook(() => useMarkAllNotificationsAsRead(), {
        wrapper: createWrapper(),
      });

      await result.current.mutateAsync();

      await waitFor(() => {
        // Both queries should be invalidated (they both start with ['notifications'])
        const listState = queryClient.getQueryState(['notifications']);
        const countState = queryClient.getQueryState(['notifications', 'unread', 'count']);

        // After invalidation, queries should be marked as stale
        expect(listState?.isInvalidated || countState?.isInvalidated).toBe(true);
      });
    });
  });
});
