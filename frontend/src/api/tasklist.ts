import { apiClient } from './client';
import type {
  TasklistItem,
  TasklistProgress,
  CreateTasklistItemRequest,
  UpdateTasklistItemRequest,
  ReorderTasklistRequest,
  BulkCreateTasklistRequest,
} from '../types';

// Get all tasklist items for an issue
export const getTasklistItems = async (issueId: number): Promise<TasklistItem[]> => {
  const response = await apiClient.get<TasklistItem[]>(`/issues/${issueId}/tasklist`);
  return response.data;
};

// Get tasklist progress for an issue
export const getTasklistProgress = async (issueId: number): Promise<TasklistProgress> => {
  const response = await apiClient.get<TasklistProgress>(`/issues/${issueId}/tasklist/progress`);
  return response.data;
};

// Create a new tasklist item
export const createTasklistItem = async (
  issueId: number,
  data: CreateTasklistItemRequest
): Promise<TasklistItem> => {
  const response = await apiClient.post<TasklistItem>(`/issues/${issueId}/tasklist`, data);
  return response.data;
};

// Bulk create tasklist items
export const bulkCreateTasklistItems = async (
  issueId: number,
  data: BulkCreateTasklistRequest
): Promise<TasklistItem[]> => {
  const response = await apiClient.post<TasklistItem[]>(`/issues/${issueId}/tasklist/bulk`, data);
  return response.data;
};

// Update a tasklist item
export const updateTasklistItem = async (
  itemId: number,
  data: UpdateTasklistItemRequest
): Promise<TasklistItem> => {
  const response = await apiClient.put<TasklistItem>(`/tasklist/${itemId}`, data);
  return response.data;
};

// Toggle tasklist item completion
export const toggleTasklistItem = async (itemId: number): Promise<TasklistItem> => {
  const response = await apiClient.patch<TasklistItem>(`/tasklist/${itemId}/toggle`);
  return response.data;
};

// Delete a tasklist item
export const deleteTasklistItem = async (itemId: number): Promise<void> => {
  await apiClient.delete(`/tasklist/${itemId}`);
};

// Reorder tasklist items
export const reorderTasklistItems = async (
  issueId: number,
  data: ReorderTasklistRequest
): Promise<void> => {
  await apiClient.put(`/issues/${issueId}/tasklist/reorder`, data);
};
