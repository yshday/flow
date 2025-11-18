import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { milestonesApi } from '../api/milestones';
import type { CreateMilestoneRequest, UpdateMilestoneRequest } from '../types';

export function useMilestones(projectId: number) {
  return useQuery({
    queryKey: ['projects', projectId, 'milestones'],
    queryFn: () => milestonesApi.list(projectId),
    enabled: !!projectId,
  });
}

export function useMilestone(id: number, withProgress?: boolean) {
  return useQuery({
    queryKey: ['milestones', id, withProgress ? 'with-progress' : null],
    queryFn: () => milestonesApi.get(id, withProgress),
    enabled: !!id,
  });
}

export function useCreateMilestone(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateMilestoneRequest) => milestonesApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'milestones'] });
    },
  });
}

export function useUpdateMilestone(id: number, projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateMilestoneRequest) => milestonesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['milestones', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'milestones'] });
    },
  });
}

export function useDeleteMilestone(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => milestonesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'milestones'] });
    },
  });
}
