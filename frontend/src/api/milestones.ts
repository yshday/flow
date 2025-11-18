import { apiClient } from './client';
import type { Milestone, CreateMilestoneRequest, UpdateMilestoneRequest } from '../types';

export const milestonesApi = {
  list: async (projectId: number): Promise<Milestone[]> => {
    const response = await apiClient.get<Milestone[]>(`/projects/${projectId}/milestones`);
    return response.data;
  },

  get: async (id: number, withProgress?: boolean): Promise<Milestone> => {
    const response = await apiClient.get<Milestone>(`/milestones/${id}`, {
      params: withProgress ? { with_progress: 'true' } : undefined,
    });
    return response.data;
  },

  create: async (projectId: number, data: CreateMilestoneRequest): Promise<Milestone> => {
    const response = await apiClient.post<Milestone>(`/projects/${projectId}/milestones`, data);
    return response.data;
  },

  update: async (id: number, data: UpdateMilestoneRequest): Promise<Milestone> => {
    const response = await apiClient.put<Milestone>(`/milestones/${id}`, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/milestones/${id}`);
  },
};
