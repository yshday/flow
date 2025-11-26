import { apiClient } from './client';
import type {
  Issue,
  CreateIssueRequest,
  UpdateIssueRequest,
  MoveIssueRequest,
  SubtaskProgress,
  Label,
  CreateLabelRequest,
  UpdateLabelRequest,
  Comment,
  CreateCommentRequest,
  UpdateCommentRequest,
  Activity,
} from '../types';

export const issuesApi = {
  // Issues
  list: async (projectId: number, params?: Record<string, any>): Promise<Issue[]> => {
    const response = await apiClient.get<Issue[]>(`/projects/${projectId}/issues`, { params });
    return response.data;
  },

  get: async (id: number): Promise<Issue> => {
    const response = await apiClient.get<Issue>(`/issues/${id}`);
    return response.data;
  },

  getByKey: async (projectKey: string, issueNumber: number): Promise<Issue> => {
    const response = await apiClient.get<Issue>(`/issues/${projectKey}/${issueNumber}`);
    return response.data;
  },

  getByNumber: async (projectId: number, issueNumber: number): Promise<Issue> => {
    const response = await apiClient.get<Issue>(`/projects/${projectId}/issue-by-number/${issueNumber}`);
    return response.data;
  },

  create: async (projectId: number, data: CreateIssueRequest): Promise<Issue> => {
    const response = await apiClient.post<Issue>(`/projects/${projectId}/issues`, data);
    return response.data;
  },

  update: async (id: number, data: UpdateIssueRequest): Promise<Issue> => {
    const response = await apiClient.put<Issue>(`/issues/${id}`, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/issues/${id}`);
  },

  move: async (id: number, data: MoveIssueRequest): Promise<Issue> => {
    const response = await apiClient.put<Issue>(`/issues/${id}/move`, data);
    return response.data;
  },

  // Subtasks
  getSubtasks: async (issueId: number): Promise<Issue[]> => {
    const response = await apiClient.get<Issue[]>(`/issues/${issueId}/subtasks`);
    return response.data;
  },

  getSubtaskProgress: async (issueId: number): Promise<SubtaskProgress> => {
    const response = await apiClient.get<SubtaskProgress>(`/issues/${issueId}/subtasks/progress`);
    return response.data;
  },

  // Epics
  getEpics: async (projectId: number): Promise<Issue[]> => {
    const response = await apiClient.get<Issue[]>(`/projects/${projectId}/epics`);
    return response.data;
  },

  getEpicIssues: async (epicId: number): Promise<Issue[]> => {
    const response = await apiClient.get<Issue[]>(`/issues/${epicId}/epic-issues`);
    return response.data;
  },

  getEpicProgress: async (epicId: number): Promise<SubtaskProgress> => {
    const response = await apiClient.get<SubtaskProgress>(`/issues/${epicId}/epic-progress`);
    return response.data;
  },

  // Labels
  listLabels: async (projectId: number): Promise<Label[]> => {
    const response = await apiClient.get<Label[]>(`/projects/${projectId}/labels`);
    return response.data;
  },

  createLabel: async (projectId: number, data: CreateLabelRequest): Promise<Label> => {
    const response = await apiClient.post<Label>(`/projects/${projectId}/labels`, data);
    return response.data;
  },

  updateLabel: async (labelId: number, data: UpdateLabelRequest): Promise<Label> => {
    const response = await apiClient.put<Label>(`/labels/${labelId}`, data);
    return response.data;
  },

  deleteLabel: async (labelId: number): Promise<void> => {
    await apiClient.delete(`/labels/${labelId}`);
  },

  // Issue Labels
  getIssueLabels: async (issueId: number): Promise<Label[]> => {
    const response = await apiClient.get<Label[]>(`/issues/${issueId}/labels`);
    return response.data;
  },

  addLabelToIssue: async (issueId: number, labelId: number): Promise<void> => {
    await apiClient.post(`/issues/${issueId}/labels/${labelId}`);
  },

  removeLabelFromIssue: async (issueId: number, labelId: number): Promise<void> => {
    await apiClient.delete(`/issues/${issueId}/labels/${labelId}`);
  },

  // Comments
  listComments: async (issueId: number): Promise<Comment[]> => {
    const response = await apiClient.get<Comment[]>(`/issues/${issueId}/comments`);
    return response.data;
  },

  createComment: async (issueId: number, data: CreateCommentRequest): Promise<Comment> => {
    const response = await apiClient.post<Comment>(`/issues/${issueId}/comments`, data);
    return response.data;
  },

  updateComment: async (commentId: number, data: UpdateCommentRequest): Promise<Comment> => {
    const response = await apiClient.put<Comment>(`/comments/${commentId}`, data);
    return response.data;
  },

  deleteComment: async (commentId: number): Promise<void> => {
    await apiClient.delete(`/comments/${commentId}`);
  },

  // Activities
  listActivities: async (issueId: number, params?: { limit?: number; offset?: number }): Promise<Activity[]> => {
    const response = await apiClient.get<Activity[]>(`/issues/${issueId}/activities`, { params });
    return response.data;
  },
};
