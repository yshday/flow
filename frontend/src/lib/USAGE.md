# @flow/issue-tracker 사용 가이드

## 설치

```bash
npm install @flow/issue-tracker
```

## 기본 사용법

```tsx
import { FlowIssueTracker } from '@flow/issue-tracker'

function App() {
  // 호스트 앱에서 가져온 사용자 정보와 토큰
  const user = {
    id: 1,
    email: 'user@example.com',
    username: 'user',
  }

  const company = {
    id: 1,
    name: 'Company Name',
  }

  const accessToken = 'your-flow-access-token'

  return (
    <FlowIssueTracker
      config={{
        apiBaseUrl: 'http://localhost:8080/api/v1',
        debug: true,
      }}
      user={user}
      company={company}
      accessToken={accessToken}
      callbacks={{
        onIssueCreate: (issue) => console.log('Issue created:', issue),
        onIssueClick: (issue) => console.log('Issue clicked:', issue),
        onNavigate: (path) => console.log('Navigate to:', path),
      }}
    />
  )
}
```

## 개별 컴포넌트 사용

더 세밀한 제어가 필요한 경우 개별 컴포넌트를 사용할 수 있습니다:

```tsx
import {
  FlowProvider,
  ProjectList,
  KanbanBoard,
  IssueDetail,
} from '@flow/issue-tracker'

function CustomApp() {
  return (
    <FlowProvider
      config={{ apiBaseUrl: 'http://localhost:8080/api/v1' }}
      user={user}
      accessToken={accessToken}
    >
      <div className="flex">
        <aside>
          <ProjectList onProjectClick={handleProjectClick} />
        </aside>
        <main>
          {selectedProject && (
            <KanbanBoard
              projectId={selectedProject.id}
              onIssueClick={handleIssueClick}
            />
          )}
        </main>
      </div>
    </FlowProvider>
  )
}
```

## 커스텀 Hooks 사용

API 데이터에 직접 접근하려면 hooks를 사용합니다:

```tsx
import {
  FlowProvider,
  useFlowProjects,
  useFlowIssues,
  useFlowCreateIssue,
} from '@flow/issue-tracker'

function CustomProjectView() {
  const { data: projects, isLoading } = useFlowProjects()
  const { data: issues } = useFlowIssues(projectId)
  const { mutateAsync: createIssue } = useFlowCreateIssue(projectId)

  const handleCreate = async () => {
    await createIssue({
      title: 'New Issue',
      description: 'Issue description',
      priority: 'medium',
    })
  }

  // ...
}
```

## Props 설명

### FlowIssueTracker

| Prop | 타입 | 필수 | 설명 |
|------|------|------|------|
| config | FlowConfig | O | API 설정 |
| user | FlowUser | O | 사용자 정보 |
| company | FlowCompany | X | 회사 정보 |
| accessToken | string | O | Flow API 토큰 |
| initialProjectId | number | X | 초기 프로젝트 ID |
| callbacks | FlowEventCallbacks | X | 이벤트 콜백 |
| className | string | X | 커스텀 클래스명 |

### FlowConfig

```typescript
interface FlowConfig {
  apiBaseUrl: string  // Flow API 서버 URL
  debug?: boolean     // 디버그 모드
  theme?: {
    primary?: string
    secondary?: string
  }
}
```

### FlowEventCallbacks

```typescript
interface FlowEventCallbacks {
  onIssueCreate?: (issue: Issue) => void
  onIssueUpdate?: (issue: Issue) => void
  onIssueDelete?: (issueId: number) => void
  onIssueClick?: (issue: Issue) => void
  onProjectCreate?: (project: Project) => void
  onProjectClick?: (project: Project) => void
  onNavigate?: (path: string) => void
  onError?: (error: ApiError) => void
}
```

## 스타일링

패키지는 TailwindCSS 클래스를 사용합니다. 호스트 앱에 TailwindCSS가 설정되어 있어야 합니다.

```tsx
// 호스트 앱에서 TailwindCSS import
import 'tailwindcss/tailwind.css'
```

커스텀 스타일을 적용하려면 className prop을 사용하세요:

```tsx
<FlowIssueTracker
  className="custom-wrapper"
  // ...
/>
```

## TypeScript 지원

패키지는 완전한 TypeScript 지원을 제공합니다:

```typescript
import type {
  Issue,
  Project,
  IssueStatus,
  IssuePriority,
  FlowUser,
  FlowCompany,
} from '@flow/issue-tracker'
```
