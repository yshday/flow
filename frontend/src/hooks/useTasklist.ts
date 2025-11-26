import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getTasklistItems,
  getTasklistProgress,
  createTasklistItem,
  updateTasklistItem,
  toggleTasklistItem,
  deleteTasklistItem,
  reorderTasklistItems,
} from '../api/tasklist';
import type {
  TasklistItem,
  CreateTasklistItemRequest,
  UpdateTasklistItemRequest,
  ReorderTasklistRequest,
} from '../types';

export const useTasklistItems = (issueId: number) => {
  return useQuery({
    queryKey: ['tasklist', issueId],
    queryFn: () => getTasklistItems(issueId),
    enabled: !!issueId,
  });
};

export const useTasklistProgress = (issueId: number) => {
  return useQuery({
    queryKey: ['tasklist-progress', issueId],
    queryFn: () => getTasklistProgress(issueId),
    enabled: !!issueId,
  });
};

export const useCreateTasklistItem = (issueId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTasklistItemRequest) => createTasklistItem(issueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasklist', issueId] });
      queryClient.invalidateQueries({ queryKey: ['tasklist-progress', issueId] });
    },
  });
};

export const useUpdateTasklistItem = (issueId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ itemId, data }: { itemId: number; data: UpdateTasklistItemRequest }) =>
      updateTasklistItem(itemId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasklist', issueId] });
      queryClient.invalidateQueries({ queryKey: ['tasklist-progress', issueId] });
    },
  });
};

export const useToggleTasklistItem = (issueId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (itemId: number) => toggleTasklistItem(itemId),
    onMutate: async (itemId) => {
      // Optimistic update
      await queryClient.cancelQueries({ queryKey: ['tasklist', issueId] });

      const previousItems = queryClient.getQueryData<TasklistItem[]>(['tasklist', issueId]);

      if (previousItems) {
        queryClient.setQueryData<TasklistItem[]>(['tasklist', issueId], (old) =>
          old?.map((item) =>
            item.id === itemId ? { ...item, is_completed: !item.is_completed } : item
          )
        );
      }

      return { previousItems };
    },
    onError: (_err, _itemId, context) => {
      if (context?.previousItems) {
        queryClient.setQueryData(['tasklist', issueId], context.previousItems);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tasklist', issueId] });
      queryClient.invalidateQueries({ queryKey: ['tasklist-progress', issueId] });
    },
  });
};

export const useDeleteTasklistItem = (issueId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (itemId: number) => deleteTasklistItem(itemId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasklist', issueId] });
      queryClient.invalidateQueries({ queryKey: ['tasklist-progress', issueId] });
    },
  });
};

export const useReorderTasklistItems = (issueId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ReorderTasklistRequest) => reorderTasklistItems(issueId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasklist', issueId] });
    },
  });
};
