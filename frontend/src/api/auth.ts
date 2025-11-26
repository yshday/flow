import { apiClient } from './client';
import type { AuthResponse, LoginRequest, RegisterRequest, User, ProjectMembership } from '../types';

export const authApi = {
  register: async (data: RegisterRequest): Promise<User> => {
    const response = await apiClient.post<User>('/auth/register', data);
    return response.data;
  },

  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/login', data);
    return response.data;
  },

  refreshToken: async (refreshToken: string): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get<User>('/auth/me');
    return response.data;
  },

  getUserMemberships: async (): Promise<ProjectMembership[]> => {
    const response = await apiClient.get<ProjectMembership[]>('/users/me/memberships');
    return response.data;
  },

  searchUsers: async (query: string, limit?: number): Promise<User[]> => {
    const params = new URLSearchParams({ q: query });
    if (limit) {
      params.append('limit', limit.toString());
    }
    const response = await apiClient.get<User[]>(`/users/search?${params.toString()}`);
    return response.data;
  },

  logout: async (): Promise<void> => {
    // Clear tokens from localStorage
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  },
};
