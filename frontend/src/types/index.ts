// User types
export interface User {
  id: number;
  email: string;
  username: string;
  avatar_url?: string;
  created_at: string;
  updated_at: string;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

// Project types
export interface Project {
  id: number;
  name: string;
  key: string;
  description?: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  name: string;
  key: string;
  description?: string;
  template_id?: number;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

// Board Column types
export interface BoardColumn {
  id: number;
  project_id: number;
  name: string;
  position: number;
  created_at: string;
}

export interface CreateColumnRequest {
  name: string;
  position: number;
}

export interface UpdateColumnRequest {
  name?: string;
  position?: number;
}

// Issue types
export type IssueStatus = 'open' | 'in_progress' | 'closed';
export type IssuePriority = 'low' | 'medium' | 'high' | 'urgent';
export type IssueType = 'bug' | 'improvement' | 'epic' | 'feature' | 'task' | 'subtask';

export interface Issue {
  id: number;
  project_id: number;
  issue_number: number;
  title: string;
  description?: string;
  status: IssueStatus;
  priority: IssuePriority;
  issue_type: IssueType;
  parent_issue_id?: number;
  epic_id?: number;
  column_id?: number;
  column_position?: number;
  assignee_id?: number;
  reporter_id: number;
  milestone_id?: number;
  version: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  // Related entities
  parent_issue?: Issue;
  epic?: Issue;
  subtasks?: Issue[];
  epic_issues?: Issue[];
}

export interface CreateIssueRequest {
  title: string;
  description?: string;
  priority?: IssuePriority;
  issue_type?: IssueType;
  parent_issue_id?: number;
  epic_id?: number;
  column_id?: number;
  assignee_id?: number;
  milestone_id?: number;
  label_ids?: number[];
}

export interface UpdateIssueRequest {
  title?: string;
  description?: string;
  status?: IssueStatus;
  priority?: IssuePriority;
  issue_type?: IssueType;
  epic_id?: number;
  assignee_id?: number;
  milestone_id?: number;
}

export interface SubtaskProgress {
  total: number;
  completed: number;
}

export interface MoveIssueRequest {
  column_id: number;
  version: number;
  position?: number;
}

// Label types
export interface Label {
  id: number;
  project_id: number;
  name: string;
  color: string;
  created_at: string;
}

export interface CreateLabelRequest {
  name: string;
  color: string;
}

export interface UpdateLabelRequest {
  name?: string;
  color?: string;
}

// Milestone types
export type MilestoneStatus = 'open' | 'closed';

export interface Milestone {
  id: number;
  project_id: number;
  title: string;
  description?: string;
  due_date?: string;
  status: MilestoneStatus;
  total_issues?: number;
  closed_issues?: number;
  progress?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateMilestoneRequest {
  title: string;
  description?: string;
  due_date?: string;
}

export interface UpdateMilestoneRequest {
  title?: string;
  description?: string;
  due_date?: string;
  status?: MilestoneStatus;
}

// Comment types
export interface Comment {
  id: number;
  issue_id: number;
  user_id: number;
  content: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCommentRequest {
  content: string;
}

export interface UpdateCommentRequest {
  content: string;
}

// Project Member types
export type ProjectRole = 'owner' | 'admin' | 'member' | 'viewer';

export interface ProjectMember {
  project_id: number;
  user_id: number;
  role: ProjectRole;
  joined_at: string;
  invited_by?: number;
  user?: User;
  project?: Project;
}

export interface ProjectMembership {
  project_id: number;
  user_id: number;
  role: ProjectRole;
  joined_at: string;
  invited_by?: number;
  project: Project;
}

export interface AddMemberRequest {
  user_id: number;
  role: ProjectRole;
}

export interface UpdateMemberRoleRequest {
  role: ProjectRole;
}

// Activity types
export type ActivityAction = 'created' | 'updated' | 'deleted' | 'moved' | 'added' | 'removed';
export type EntityType = 'issue' | 'comment' | 'label' | 'member' | 'project' | 'board';

export interface Activity {
  id: number;
  project_id?: number;
  issue_id?: number;
  user_id: number;
  action: ActivityAction;
  entity_type: EntityType;
  entity_id?: number;
  field_name?: string;
  old_value?: string;
  new_value?: string;
  created_at: string;
  user?: User;
}

// Pagination types
export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

// Error types
export interface ApiError {
  code: string;
  message: string;
  details?: Array<{ field: string; message: string }>;
  request_id?: string;
}

export interface ApiErrorResponse {
  error: ApiError;
}

// Notification types
export type NotificationAction =
  | 'created'
  | 'updated'
  | 'deleted'
  | 'assigned'
  | 'commented'
  | 'mentioned'
  | 'added'
  | 'removed';

export type NotificationEntityType =
  | 'issue'
  | 'comment'
  | 'project'
  | 'label'
  | 'milestone'
  | 'member';

export interface Notification {
  id: number;
  user_id: number;
  action: NotificationAction;
  entity_type: NotificationEntityType;
  entity_id: number;
  actor_id: number;
  message: string;
  is_read: boolean;
  created_at: string;
}

export interface MarkNotificationsAsReadRequest {
  notification_ids: number[];
}

// Tasklist types
export interface TasklistItem {
  id: number;
  issue_id: number;
  content: string;
  is_completed: boolean;
  position: number;
  completed_at?: string;
  completed_by?: number;
  completed_by_user?: User;
  created_at: string;
  updated_at: string;
}

export interface TasklistProgress {
  total: number;
  completed: number;
  pending: number;
  percent: number;
}

export interface CreateTasklistItemRequest {
  content: string;
  position?: number;
}

export interface UpdateTasklistItemRequest {
  content?: string;
  is_completed?: boolean;
  position?: number;
}

export interface ReorderTasklistRequest {
  item_ids: number[];
}

export interface BulkCreateTasklistRequest {
  items: CreateTasklistItemRequest[];
}

// Template types

// Project Template Config
export interface ColumnConfig {
  name: string;
  position: number;
  wip_limit?: number;
}

export interface LabelConfig {
  name: string;
  color: string;
  description?: string;
}

export interface MilestoneConfig {
  title: string;
  description?: string;
}

export interface ProjectTemplateConfig {
  columns: ColumnConfig[];
  labels: LabelConfig[];
  milestones?: MilestoneConfig[];
}

export interface ProjectTemplate {
  id: number;
  name: string;
  description?: string;
  is_system: boolean;
  created_by?: number;
  config: ProjectTemplateConfig;
  created_at: string;
  updated_at: string;
}

// Issue Template
export interface IssueTemplate {
  id: number;
  project_id: number;
  name: string;
  description?: string;
  content: string;
  default_priority: IssuePriority;
  default_labels: number[];
  position: number;
  is_active: boolean;
  created_by?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateIssueTemplateRequest {
  name: string;
  description?: string;
  content?: string;
  default_priority?: IssuePriority;
  default_labels?: number[];
  position?: number;
}

export interface UpdateIssueTemplateRequest {
  name?: string;
  description?: string;
  content?: string;
  default_priority?: IssuePriority;
  default_labels?: number[];
  position?: number;
  is_active?: boolean;
}

// Extended CreateProjectRequest with template
export interface CreateProjectWithTemplateRequest {
  name: string;
  key: string;
  description?: string;
  template_id?: number;
}
