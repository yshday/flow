import { apiClient } from './client';
import type { Notification, MarkNotificationsAsReadRequest } from '../types';

export const notificationsApi = {
  // Get all notifications
  list: async (params?: { unread?: boolean; limit?: number; offset?: number }): Promise<Notification[]> => {
    const response = await apiClient.get<Notification[]>('/notifications', { params });
    return response.data;
  },

  // Get unread notifications count
  getUnreadCount: async (): Promise<{ unread_count: number }> => {
    const response = await apiClient.get<{ unread_count: number }>('/notifications/unread/count');
    return response.data;
  },

  // Mark notifications as read
  markAsRead: async (data: MarkNotificationsAsReadRequest): Promise<void> => {
    await apiClient.put('/notifications/read', data);
  },

  // Mark all notifications as read
  markAllAsRead: async (): Promise<void> => {
    await apiClient.put('/notifications/read/all');
  },
};
