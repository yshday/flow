/**
 * Flow API Functions
 * 패키지 내부에서 사용하는 API 함수들
 */

import { getFlowClient } from './flowClient'
import type {
  Project,
  Issue,
  BoardColumn,
  Label,
  Comment,
  CreateIssueRequest,
  UpdateIssueRequest,
  ProjectMember,
  AddMemberRequest,
  UpdateMemberRoleRequest,
  Milestone,
  CreateMilestoneRequest,
  UpdateMilestoneRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
  User,
  ProjectTemplate,
  IssueTemplate,
  CreateIssueTemplateRequest,
  UpdateIssueTemplateRequest,
} from '../types'

// === Projects API ===

export const flowProjectsApi = {
  list: async (): Promise<Project[]> => {
    const { data } = await getFlowClient().get('/projects')
    return data
  },

  get: async (id: number): Promise<Project> => {
    const { data } = await getFlowClient().get(`/projects/${id}`)
    return data
  },

  listColumns: async (projectId: number): Promise<BoardColumn[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/board`)
    return data
  },
}

// === Issues API ===

export const flowIssuesApi = {
  list: async (projectId: number, params?: Record<string, any>): Promise<Issue[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/issues`, { params })
    return data
  },

  get: async (id: number): Promise<Issue> => {
    const { data } = await getFlowClient().get(`/issues/${id}`)
    return data
  },

  create: async (projectId: number, issue: CreateIssueRequest): Promise<Issue> => {
    const { data } = await getFlowClient().post(`/projects/${projectId}/issues`, issue)
    return data
  },

  update: async (id: number, issue: UpdateIssueRequest): Promise<Issue> => {
    const { data } = await getFlowClient().put(`/issues/${id}`, issue)
    return data
  },

  move: async (
    id: number,
    moveData: { column_id: number; version: number; position?: number; status?: string }
  ): Promise<Issue> => {
    const { data } = await getFlowClient().put(`/issues/${id}/move`, moveData)
    return data
  },

  delete: async (id: number): Promise<void> => {
    await getFlowClient().delete(`/issues/${id}`)
  },

  // Labels
  listLabels: async (projectId: number): Promise<Label[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/labels`)
    return data
  },

  // Comments
  listComments: async (issueId: number): Promise<Comment[]> => {
    const { data } = await getFlowClient().get(`/issues/${issueId}/comments`)
    return data
  },

  createComment: async (issueId: number, content: string): Promise<Comment> => {
    const { data } = await getFlowClient().post(`/issues/${issueId}/comments`, { content })
    return data
  },

  // Epics
  getEpics: async (projectId: number): Promise<Issue[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/epics`)
    return data
  },

  // Labels CRUD
  createLabel: async (projectId: number, label: CreateLabelRequest): Promise<Label> => {
    const { data } = await getFlowClient().post(`/projects/${projectId}/labels`, label)
    return data
  },

  updateLabel: async (labelId: number, label: UpdateLabelRequest): Promise<Label> => {
    const { data } = await getFlowClient().put(`/labels/${labelId}`, label)
    return data
  },

  deleteLabel: async (labelId: number): Promise<void> => {
    await getFlowClient().delete(`/labels/${labelId}`)
  },
}

// === Project Members API ===

export const flowMembersApi = {
  list: async (projectId: number): Promise<ProjectMember[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/members`)
    return data
  },

  add: async (projectId: number, member: AddMemberRequest): Promise<ProjectMember> => {
    const { data } = await getFlowClient().post(`/projects/${projectId}/members`, member)
    return data
  },

  updateRole: async (projectId: number, userId: number, role: UpdateMemberRoleRequest): Promise<ProjectMember> => {
    const { data } = await getFlowClient().put(`/projects/${projectId}/members/${userId}`, role)
    return data
  },

  remove: async (projectId: number, userId: number): Promise<void> => {
    await getFlowClient().delete(`/projects/${projectId}/members/${userId}`)
  },
}

// === Milestones API ===

export const flowMilestonesApi = {
  list: async (projectId: number): Promise<Milestone[]> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/milestones`)
    return data
  },

  get: async (milestoneId: number): Promise<Milestone> => {
    const { data } = await getFlowClient().get(`/milestones/${milestoneId}`)
    return data
  },

  create: async (projectId: number, milestone: CreateMilestoneRequest): Promise<Milestone> => {
    const { data } = await getFlowClient().post(`/projects/${projectId}/milestones`, milestone)
    return data
  },

  update: async (milestoneId: number, milestone: UpdateMilestoneRequest): Promise<Milestone> => {
    const { data } = await getFlowClient().put(`/milestones/${milestoneId}`, milestone)
    return data
  },

  delete: async (milestoneId: number): Promise<void> => {
    await getFlowClient().delete(`/milestones/${milestoneId}`)
  },
}

// === Users API ===

export const flowUsersApi = {
  search: async (query: string): Promise<User[]> => {
    const { data } = await getFlowClient().get('/users/search', { params: { q: query } })
    return data
  },
}

// === Templates API ===

export const flowTemplatesApi = {
  // Project Templates
  listProjectTemplates: async (): Promise<ProjectTemplate[]> => {
    const { data } = await getFlowClient().get('/templates/projects')
    return data
  },

  getProjectTemplate: async (id: number): Promise<ProjectTemplate> => {
    const { data } = await getFlowClient().get(`/templates/projects/${id}`)
    return data
  },

  // Issue Templates
  listIssueTemplates: async (projectId: number, activeOnly?: boolean): Promise<IssueTemplate[]> => {
    const params = activeOnly ? { active: 'true' } : undefined
    const { data } = await getFlowClient().get(`/projects/${projectId}/templates/issues`, { params })
    return data
  },

  getIssueTemplate: async (projectId: number, templateId: number): Promise<IssueTemplate> => {
    const { data } = await getFlowClient().get(`/projects/${projectId}/templates/issues/${templateId}`)
    return data
  },

  createIssueTemplate: async (projectId: number, template: CreateIssueTemplateRequest): Promise<IssueTemplate> => {
    const { data } = await getFlowClient().post(`/projects/${projectId}/templates/issues`, template)
    return data
  },

  updateIssueTemplate: async (projectId: number, templateId: number, template: UpdateIssueTemplateRequest): Promise<IssueTemplate> => {
    const { data } = await getFlowClient().put(`/projects/${projectId}/templates/issues/${templateId}`, template)
    return data
  },

  deleteIssueTemplate: async (projectId: number, templateId: number): Promise<void> => {
    await getFlowClient().delete(`/projects/${projectId}/templates/issues/${templateId}`)
  },
}
