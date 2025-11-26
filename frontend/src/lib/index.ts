/**
 * @flow/issue-tracker
 *
 * Flow Issue Tracker - Micro Frontend Package
 * 메신저 등 호스트 앱에 임베드할 수 있는 이슈 트래커 컴포넌트
 */

// CSS (imported for bundling - users should import '@flow/issue-tracker/styles.css' separately)
import '../index.css'

// Main Component
export { FlowIssueTracker } from './components/FlowIssueTracker'
export type {
  FlowIssueTrackerProps,
  FlowConfig,
  FlowUser,
  FlowCompany,
} from './components/FlowIssueTracker'

// Individual Components (for granular usage)
export { ProjectList } from './components/ProjectList'
export type { ProjectListProps } from './components/ProjectList'

export { ProjectSidebar } from './components/ProjectSidebar'
export type { ProjectSidebarProps } from './components/ProjectSidebar'

export { IssueTreeView } from './components/IssueTreeView'
export type { IssueTreeViewProps } from './components/IssueTreeView'

export { IssueTreeItem } from './components/IssueTreeItem'
export type { IssueTreeItemProps } from './components/IssueTreeItem'

export { KanbanBoard } from './components/KanbanBoard'
export type { KanbanBoardProps } from './components/KanbanBoard'

export { IssueDetail } from './components/IssueDetail'
export type { IssueDetailProps } from './components/IssueDetail'

export { IssueCreateModal } from './components/IssueCreateModal'
export type { IssueCreateModalProps } from './components/IssueCreateModal'

export { ProjectCreateModal } from './components/ProjectCreateModal'
export type { ProjectCreateModalProps } from './components/ProjectCreateModal'

// Providers (for advanced customization)
export { FlowProvider, useFlowConfig, useFlowCallbacks } from './providers'
export type { FlowProviderProps } from './providers'

export { FlowAuthProvider } from './providers/FlowAuthProvider'
export type { FlowAuthProviderProps } from './providers/FlowAuthProvider'

// Types
export type {
  Project,
  Issue,
  IssueStatus,
  IssuePriority,
  IssueType,
  Comment,
  Label,
  BoardColumn,
  User,
  CreateIssueRequest,
  UpdateIssueRequest,
  FlowEventCallbacks,
  PaginatedResponse,
  ApiError,
  // Member search integration
  MemberSearchResult,
  MemberSearchCallback,
} from './types'

// Hooks (for custom implementations)
export {
  useFlowProjects,
  useFlowProject,
  useFlowBoardColumns,
  useFlowIssues,
  useFlowIssue,
  useFlowCreateIssue,
  useFlowUpdateIssue,
  useFlowMoveIssue,
  useFlowDeleteIssue,
  useFlowEpics,
  useFlowLabels,
  useFlowComments,
  useFlowCreateComment,
  // Auth Bridge
  useFlowAuth,
} from './hooks'

// API (for advanced usage)
export { initFlowClient, getFlowClient, updateFlowToken } from './api'
