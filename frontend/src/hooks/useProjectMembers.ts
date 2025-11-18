import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '../api/projects';
import type { ProjectRole } from '../types';

// 프로젝트 멤버 목록 조회
export function useProjectMembers(projectId: number) {
  return useQuery({
    queryKey: ['projectMembers', projectId],
    queryFn: () => projectsApi.listMembers(projectId),
    enabled: !!projectId,
  });
}

// 멤버 추가
export function useAddMember(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { user_id: number; role: ProjectRole }) =>
      projectsApi.addMember(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projectMembers', projectId] });
    },
  });
}

// 멤버 역할 변경
export function useUpdateMemberRole(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { userId: number; role: ProjectRole }) =>
      projectsApi.updateMemberRole(projectId, data.userId, { role: data.role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projectMembers', projectId] });
    },
  });
}

// 멤버 제거
export function useRemoveMember(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: number) => projectsApi.removeMember(projectId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projectMembers', projectId] });
    },
  });
}
