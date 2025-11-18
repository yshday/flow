import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '../api/projects';
import type { CreateProjectRequest, UpdateProjectRequest } from '../types';

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list(),
  });
}

export function useProject(id: number) {
  return useQuery({
    queryKey: ['projects', id],
    queryFn: () => projectsApi.get(id),
    enabled: !!id,
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateProjectRequest) => projectsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}

export function useUpdateProject(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateProjectRequest) => projectsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projects', id] });
    },
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => projectsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}

// Board Columns
export function useBoardColumns(projectId: number) {
  return useQuery({
    queryKey: ['projects', projectId, 'board'],
    queryFn: () => projectsApi.listColumns(projectId),
    enabled: !!projectId,
  });
}

export function useCreateColumn(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; position: number }) =>
      projectsApi.createColumn(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'board'] });
    },
  });
}

export function useUpdateColumn(columnId: number, projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; position: number }) =>
      projectsApi.updateColumn(columnId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'board'] });
    },
  });
}

export function useDeleteColumn(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (columnId: number) => projectsApi.deleteColumn(columnId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'board'] });
    },
  });
}
