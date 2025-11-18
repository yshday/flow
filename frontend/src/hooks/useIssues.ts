import { useMutation, useQuery, useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { issuesApi } from '../api/issues';
import type { CreateIssueRequest, UpdateIssueRequest, MoveIssueRequest, Issue } from '../types';

const ISSUES_PER_PAGE = 20;

export function useIssues(projectId: number, params?: Record<string, any>) {
  return useQuery({
    queryKey: ['projects', projectId, 'issues', params],
    queryFn: () => issuesApi.list(projectId, params),
    enabled: !!projectId,
  });
}

export function useInfiniteIssues(projectId: number, filters?: Record<string, any>) {
  return useInfiniteQuery({
    queryKey: ['projects', projectId, 'issues', 'infinite', filters],
    queryFn: ({ pageParam = 0 }) =>
      issuesApi.list(projectId, {
        ...filters,
        limit: ISSUES_PER_PAGE,
        offset: pageParam,
      }),
    getNextPageParam: (lastPage, allPages) => {
      // If the last page has fewer items than the page size, there are no more pages
      if (!lastPage || lastPage.length < ISSUES_PER_PAGE) {
        return undefined;
      }
      // Calculate the next offset
      return allPages.length * ISSUES_PER_PAGE;
    },
    initialPageParam: 0,
    enabled: !!projectId,
  });
}

export function useIssue(id: number) {
  return useQuery({
    queryKey: ['issues', id],
    queryFn: () => issuesApi.get(id),
    enabled: !!id,
  });
}

export function useCreateIssue(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateIssueRequest) => issuesApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'issues'] });
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'board'] });
    },
  });
}

export function useUpdateIssue(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateIssueRequest) => issuesApi.update(id, data),
    onSuccess: (updatedIssue) => {
      queryClient.invalidateQueries({ queryKey: ['issues', id] });
      queryClient.invalidateQueries({ queryKey: ['projects', updatedIssue.project_id, 'issues'] });
    },
  });
}

export function useMoveIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: MoveIssueRequest }) =>
      issuesApi.move(id, data),
    onSuccess: (updatedIssue) => {
      // Invalidate and refetch to ensure UI updates immediately
      queryClient.invalidateQueries({ queryKey: ['issues', updatedIssue.id], refetchType: 'active' });
      queryClient.invalidateQueries({ queryKey: ['projects', updatedIssue.project_id, 'issues'], refetchType: 'active' });
      queryClient.invalidateQueries({ queryKey: ['projects', updatedIssue.project_id, 'board'], refetchType: 'active' });
    },
  });
}

export function useDeleteIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => issuesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
}

// Comments
export function useComments(issueId: number) {
  return useQuery({
    queryKey: ['issues', issueId, 'comments'],
    queryFn: () => issuesApi.listComments(issueId),
    enabled: !!issueId,
  });
}

export function useCreateComment(issueId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (content: string) => issuesApi.createComment(issueId, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues', issueId, 'comments'] });
    },
  });
}

// Activities
export function useActivities(issueId: number, params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['issues', issueId, 'activities', params],
    queryFn: () => issuesApi.listActivities(issueId, params),
    enabled: !!issueId,
  });
}

// Labels
export function useLabels(projectId: number) {
  return useQuery({
    queryKey: ['projects', projectId, 'labels'],
    queryFn: () => issuesApi.listLabels(projectId),
    enabled: !!projectId,
  });
}

export function useIssueLabels(issueId: number) {
  return useQuery({
    queryKey: ['issues', issueId, 'labels'],
    queryFn: () => issuesApi.getIssueLabels(issueId),
    enabled: !!issueId,
  });
}

export function useAddLabelToIssue(issueId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (labelId: number) => issuesApi.addLabelToIssue(issueId, labelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues', issueId, 'labels'] });
      queryClient.invalidateQueries({ queryKey: ['issues', issueId] });
    },
  });
}

export function useRemoveLabelFromIssue(issueId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (labelId: number) => issuesApi.removeLabelFromIssue(issueId, labelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues', issueId, 'labels'] });
      queryClient.invalidateQueries({ queryKey: ['issues', issueId] });
    },
  });
}

// Label Management
export function useCreateLabel(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; color: string; description?: string }) =>
      issuesApi.createLabel(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'labels'] });
    },
  });
}

export function useUpdateLabel(labelId: number, projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name?: string; color?: string; description?: string }) =>
      issuesApi.updateLabel(labelId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'labels'] });
    },
  });
}

export function useDeleteLabel(projectId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (labelId: number) => issuesApi.deleteLabel(labelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'labels'] });
    },
  });
}
