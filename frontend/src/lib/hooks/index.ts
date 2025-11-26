/**
 * Flow Hooks
 * 패키지에서 export되는 React Query 기반 hooks
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { flowProjectsApi, flowIssuesApi, flowMembersApi, flowMilestonesApi, flowUsersApi } from '../api'
import type {
  CreateIssueRequest,
  UpdateIssueRequest,
  AddMemberRequest,
  ProjectRole,
  CreateMilestoneRequest,
  UpdateMilestoneRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
} from '../types'

// Re-export Auth Bridge Hook
export { useFlowAuth } from './useFlowAuth'

// === Project Hooks ===

export function useFlowProjects() {
  return useQuery({
    queryKey: ['flow', 'projects'],
    queryFn: () => flowProjectsApi.list(),
  })
}

export function useFlowProject(id: number) {
  return useQuery({
    queryKey: ['flow', 'projects', id],
    queryFn: () => flowProjectsApi.get(id),
    enabled: !!id,
  })
}

export function useFlowBoardColumns(projectId: number) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'columns'],
    queryFn: () => flowProjectsApi.listColumns(projectId),
    enabled: !!projectId,
  })
}

// === Issue Hooks ===

export function useFlowIssues(projectId: number, params?: Record<string, any>) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'issues', params],
    queryFn: () => flowIssuesApi.list(projectId, params),
    enabled: !!projectId,
  })
}

export function useFlowIssue(id: number) {
  return useQuery({
    queryKey: ['flow', 'issues', id],
    queryFn: () => flowIssuesApi.get(id),
    enabled: !!id,
  })
}

export function useFlowCreateIssue(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateIssueRequest) => flowIssuesApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'issues'] })
    },
  })
}

export function useFlowUpdateIssue(id: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateIssueRequest) => flowIssuesApi.update(id, data),
    onSuccess: (updatedIssue) => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'issues', id] })
      queryClient.invalidateQueries({
        queryKey: ['flow', 'projects', updatedIssue.project_id, 'issues'],
      })
    },
  })
}

export function useFlowMoveIssue() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number
      data: { column_id: number; version: number; position?: number; status?: string }
    }) => flowIssuesApi.move(id, data),
    onSuccess: (updatedIssue) => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'issues', updatedIssue.id] })
      queryClient.invalidateQueries({
        queryKey: ['flow', 'projects', updatedIssue.project_id, 'issues'],
      })
    },
  })
}

export function useFlowDeleteIssue() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => flowIssuesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow'] })
    },
  })
}

// === Epic Hooks ===

export function useFlowEpics(projectId: number) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'epics'],
    queryFn: () => flowIssuesApi.getEpics(projectId),
    enabled: !!projectId,
  })
}

// === Label Hooks ===

export function useFlowLabels(projectId: number) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'labels'],
    queryFn: () => flowIssuesApi.listLabels(projectId),
    enabled: !!projectId,
  })
}

// === Comment Hooks ===

export function useFlowComments(issueId: number) {
  return useQuery({
    queryKey: ['flow', 'issues', issueId, 'comments'],
    queryFn: () => flowIssuesApi.listComments(issueId),
    enabled: !!issueId,
  })
}

export function useFlowCreateComment(issueId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (content: string) => flowIssuesApi.createComment(issueId, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'issues', issueId, 'comments'] })
    },
  })
}

// === Label CRUD Hooks ===

export function useFlowCreateLabel(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateLabelRequest) => flowIssuesApi.createLabel(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'labels'] })
    },
  })
}

export function useFlowUpdateLabel(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ labelId, data }: { labelId: number; data: UpdateLabelRequest }) =>
      flowIssuesApi.updateLabel(labelId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'labels'] })
    },
  })
}

export function useFlowDeleteLabel(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (labelId: number) => flowIssuesApi.deleteLabel(labelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'labels'] })
    },
  })
}

// === Project Member Hooks ===

export function useFlowProjectMembers(projectId: number) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'members'],
    queryFn: () => flowMembersApi.list(projectId),
    enabled: !!projectId,
  })
}

export function useFlowAddMember(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: AddMemberRequest) => flowMembersApi.add(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'members'] })
    },
  })
}

export function useFlowUpdateMemberRole(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: ProjectRole }) =>
      flowMembersApi.updateRole(projectId, userId, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'members'] })
    },
  })
}

export function useFlowRemoveMember(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (userId: number) => flowMembersApi.remove(projectId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'members'] })
    },
  })
}

// === Milestone Hooks ===

export function useFlowMilestones(projectId: number) {
  return useQuery({
    queryKey: ['flow', 'projects', projectId, 'milestones'],
    queryFn: () => flowMilestonesApi.list(projectId),
    enabled: !!projectId,
  })
}

export function useFlowCreateMilestone(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateMilestoneRequest) => flowMilestonesApi.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'milestones'] })
    },
  })
}

export function useFlowUpdateMilestone(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ milestoneId, data }: { milestoneId: number; data: UpdateMilestoneRequest }) =>
      flowMilestonesApi.update(milestoneId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'milestones'] })
    },
  })
}

export function useFlowDeleteMilestone(projectId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (milestoneId: number) => flowMilestonesApi.delete(milestoneId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flow', 'projects', projectId, 'milestones'] })
    },
  })
}

// === User Search Hook ===

export function useFlowSearchUsers(query: string) {
  return useQuery({
    queryKey: ['flow', 'users', 'search', query],
    queryFn: () => flowUsersApi.search(query),
    enabled: query.length >= 2,
  })
}
