import { apiClient } from './client';
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  BoardColumn,
  CreateColumnRequest,
  UpdateColumnRequest,
  ProjectMember,
  AddMemberRequest,
  UpdateMemberRoleRequest,
} from '../types';

export const projectsApi = {
  // Projects
  list: async (): Promise<Project[]> => {
    const response = await apiClient.get<Project[]>('/projects');
    return response.data;
  },

  get: async (id: number): Promise<Project> => {
    const response = await apiClient.get<Project>(`/projects/${id}`);
    return response.data;
  },

  create: async (data: CreateProjectRequest): Promise<Project> => {
    const response = await apiClient.post<Project>('/projects', data);
    return response.data;
  },

  update: async (id: number, data: UpdateProjectRequest): Promise<Project> => {
    const response = await apiClient.put<Project>(`/projects/${id}`, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/projects/${id}`);
  },

  // Board Columns
  listColumns: async (projectId: number): Promise<BoardColumn[]> => {
    const response = await apiClient.get<BoardColumn[]>(`/projects/${projectId}/board`);
    return response.data;
  },

  createColumn: async (projectId: number, data: CreateColumnRequest): Promise<BoardColumn> => {
    const response = await apiClient.post<BoardColumn>(
      `/projects/${projectId}/board/columns`,
      data
    );
    return response.data;
  },

  updateColumn: async (columnId: number, data: UpdateColumnRequest): Promise<BoardColumn> => {
    const response = await apiClient.put<BoardColumn>(`/board/columns/${columnId}`, data);
    return response.data;
  },

  deleteColumn: async (columnId: number): Promise<void> => {
    await apiClient.delete(`/board/columns/${columnId}`);
  },

  // Project Members
  listMembers: async (projectId: number): Promise<ProjectMember[]> => {
    const response = await apiClient.get<ProjectMember[]>(`/projects/${projectId}/members`);
    return response.data;
  },

  addMember: async (projectId: number, data: AddMemberRequest): Promise<void> => {
    await apiClient.post(`/projects/${projectId}/members`, data);
  },

  updateMemberRole: async (
    projectId: number,
    userId: number,
    data: UpdateMemberRoleRequest
  ): Promise<void> => {
    await apiClient.put(`/projects/${projectId}/members/${userId}`, data);
  },

  removeMember: async (projectId: number, userId: number): Promise<void> => {
    await apiClient.delete(`/projects/${projectId}/members/${userId}`);
  },
};
