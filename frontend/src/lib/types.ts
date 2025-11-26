/**
 * Flow Issue Tracker - Type Definitions
 * 메신저 통합을 위한 타입 정의
 */

// === Issue Types ===

export type IssueStatus = 'open' | 'in_progress' | 'closed'
export type IssuePriority = 'low' | 'medium' | 'high' | 'urgent'
export type IssueType = 'bug' | 'feature' | 'task' | 'epic' | 'subtask' | 'improvement'

export interface Issue {
  id: number
  project_id: number
  issue_number: number
  title: string
  description?: string
  description_html?: string
  status: IssueStatus
  priority: IssuePriority
  issue_type: IssueType
  assignee_id?: number
  assignee?: User
  reporter_id: number
  reporter?: User
  column_id?: number
  column?: BoardColumn
  milestone_id?: number
  parent_issue_id?: number
  epic_id?: number
  labels?: Label[]
  version: number
  is_pinned: boolean
  created_at: string
  updated_at: string
  closed_at?: string
  // For tree view (populated on client-side)
  epic_issues?: Issue[]
  subtasks?: Issue[]
}

export interface CreateIssueRequest {
  title: string
  description?: string
  priority?: IssuePriority
  issue_type?: IssueType
  assignee_id?: number
  column_id?: number
  milestone_id?: number
  parent_issue_id?: number
  epic_id?: number
  label_ids?: number[]
}

export interface UpdateIssueRequest {
  title?: string
  description?: string
  status?: IssueStatus
  priority?: IssuePriority
  assignee_id?: number | null
  column_id?: number | null
  milestone_id?: number | null
  version: number
}

// === Project Types ===

export interface Project {
  id: number
  name: string
  key: string
  description?: string
  owner_id: number
  owner?: User
  created_at: string
  updated_at: string
}

export interface CreateProjectRequest {
  name: string
  key: string
  description?: string
  template_id?: number
}

// === Board Types ===

export interface BoardColumn {
  id: number
  project_id: number
  name: string
  position: number
  color?: string
  created_at: string
  updated_at: string
}

// === User Types ===

export interface User {
  id: number
  email: string
  username: string
  name?: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

// === Label Types ===

export interface Label {
  id: number
  project_id: number
  name: string
  color: string
  description?: string
  created_at: string
  updated_at: string
}

// === Comment Types ===

export interface Comment {
  id: number
  issue_id: number
  user_id: number
  user?: User
  content: string
  content_html?: string
  created_at: string
  updated_at: string
}

// === Project Member Types ===

export type ProjectRole = 'owner' | 'admin' | 'member' | 'viewer'

export interface ProjectMember {
  user_id: number
  project_id: number
  role: ProjectRole
  user?: User
  created_at: string
  updated_at: string
}

export interface AddMemberRequest {
  user_id: number
  role: ProjectRole
}

export interface UpdateMemberRoleRequest {
  role: ProjectRole
}

// === Milestone Types ===

export type MilestoneStatus = 'open' | 'closed'

export interface Milestone {
  id: number
  project_id: number
  title: string
  description?: string
  status: MilestoneStatus
  due_date?: string
  created_at: string
  updated_at: string
}

export interface CreateMilestoneRequest {
  title: string
  description?: string
  due_date?: string
}

export interface UpdateMilestoneRequest {
  title?: string
  description?: string
  status?: MilestoneStatus
  due_date?: string
}

// === Label CRUD Types ===

export interface CreateLabelRequest {
  name: string
  color: string
  description?: string
}

export interface UpdateLabelRequest {
  name?: string
  color?: string
  description?: string
}

// === API Response Types ===

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
  has_more: boolean
}

export interface ApiError {
  error: {
    message: string
    code?: string
  }
}

// === Event Callback Types ===

export type IssueEventType = 'created' | 'updated' | 'deleted' | 'moved' | 'assigned'
export type ProjectEventType = 'created' | 'updated' | 'deleted'

/** 멤버 검색 결과 타입 (호스트 앱에서 반환) */
export interface MemberSearchResult {
  id: number
  name: string
  email?: string
  avatar_url?: string
  department?: string
  rank?: string
}

/** 멤버 검색 콜백 타입 */
export type MemberSearchCallback = (query: string) => Promise<MemberSearchResult[]>

export interface FlowEventCallbacks {
  onIssueCreate?: (issue: Issue) => void
  onIssueUpdate?: (issue: Issue) => void
  onIssueDelete?: (issueId: number) => void
  onIssueClick?: (issue: Issue) => void
  onProjectCreate?: (project: Project) => void
  onProjectClick?: (project: Project) => void
  onNavigate?: (path: string) => void
  onError?: (error: ApiError) => void
  /** 호스트 앱의 조직도에서 멤버 검색 */
  onSearchMembers?: MemberSearchCallback
}

// === Template Types ===

export interface ColumnConfig {
  name: string
  position: number
  wip_limit?: number
}

export interface LabelConfig {
  name: string
  color: string
  description?: string
}

export interface MilestoneConfig {
  title: string
  description?: string
}

export interface ProjectTemplateConfig {
  columns: ColumnConfig[]
  labels: LabelConfig[]
  milestones?: MilestoneConfig[]
}

export interface ProjectTemplate {
  id: number
  name: string
  description?: string
  is_system: boolean
  created_by?: number
  config: ProjectTemplateConfig
  created_at: string
  updated_at: string
}

export interface IssueTemplate {
  id: number
  project_id: number
  name: string
  description?: string
  content: string
  default_priority: IssuePriority
  default_labels: number[]
  position: number
  is_active: boolean
  created_by?: number
  created_at: string
  updated_at: string
}

export interface CreateIssueTemplateRequest {
  name: string
  description?: string
  content?: string
  default_priority?: IssuePriority
  default_labels?: number[]
  position?: number
}

export interface UpdateIssueTemplateRequest {
  name?: string
  description?: string
  content?: string
  default_priority?: IssuePriority
  default_labels?: number[]
  position?: number
  is_active?: boolean
}
