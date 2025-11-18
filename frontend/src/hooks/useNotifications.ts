import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsApi } from '../api/notifications';
import type { MarkNotificationsAsReadRequest } from '../types';

// Get all notifications
export function useNotifications(params?: { unread?: boolean; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['notifications', params],
    queryFn: () => notificationsApi.list(params),
    refetchInterval: 30000, // Refetch every 30 seconds
    refetchIntervalInBackground: false, // Don't refetch when tab is not active (saves battery)
  });
}

// Get unread notifications count
export function useUnreadNotificationsCount() {
  return useQuery({
    queryKey: ['notifications', 'unread', 'count'],
    queryFn: () => notificationsApi.getUnreadCount(),
    refetchInterval: 10000, // Refetch every 10 seconds
    refetchIntervalInBackground: false, // Don't refetch when tab is not active (saves battery)
  });
}

// Mark notifications as read
export function useMarkNotificationsAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: MarkNotificationsAsReadRequest) => notificationsApi.markAsRead(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}

// Mark all notifications as read
export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => notificationsApi.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
}
