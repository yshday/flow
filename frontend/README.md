# Flow - Issue Tracker Frontend

React + TypeScript + Vite 기반의 이슈 트래커 프론트엔드

## 기술 스택

- **React 18** - UI 라이브러리
- **TypeScript** - 타입 안정성
- **Vite** - 빌드 도구
- **TanStack Query (React Query)** - 서버 상태 관리
- **React Router** - 라우팅
- **Tailwind CSS** - 스타일링
- **Axios** - HTTP 클라이언트

## 주요 기능

### ✅ 완료된 기능

#### 1. 인증 시스템
- 로그인 / 회원가입
- JWT 토큰 기반 인증
- Protected Routes
- 자동 토큰 갱신

#### 2. 프로젝트 관리
- 프로젝트 생성 / 수정 / 삭제
- 프로젝트 목록 조회
- 프로젝트 상세 정보

#### 3. 이슈 관리
- 칸반 보드 (드래그 앤 드롭)
- 이슈 생성 / 수정 / 삭제
- 이슈 상태 자동 업데이트 (칸반 이동 시)
- 이슈 필터링 (상태, 우선순위, 담당자)
- 무한 스크롤 (Infinite Scroll)
- 이슈 상세 페이지

#### 4. 라벨 관리
- 라벨 생성 / 수정 / 삭제
- 이슈에 라벨 추가 / 제거
- 색상 커스터마이징

#### 5. 마일스톤 관리
- 마일스톤 생성 / 수정 / 삭제
- 이슈와 마일스톤 연결
- 진행률 표시

#### 6. 프로젝트 멤버 관리 (신규 ✨)
- 멤버 목록 조회
- 멤버 역할 변경 (Owner, Admin, Member, Viewer)
- 멤버 제거
- 멤버 추가 (구현 중)

### 🚧 개발 중인 기능

- 멤버 추가 모달
- 댓글 시스템
- 알림 시스템
- 검색 기능

## 프로젝트 구조

```
src/
├── api/              # API 클라이언트
│   ├── client.ts     # Axios 설정
│   ├── auth.ts       # 인증 API
│   ├── projects.ts   # 프로젝트 & 멤버 API
│   ├── issues.ts     # 이슈 API
│   └── milestones.ts # 마일스톤 API
├── components/       # 재사용 가능한 컴포넌트
│   ├── layout/       # 레이아웃 컴포넌트
│   ├── kanban/       # 칸반 보드
│   └── ui/           # 공통 UI 컴포넌트
├── hooks/            # Custom Hooks
│   ├── useAuth.ts
│   ├── useProjects.ts
│   ├── useIssues.ts
│   ├── useProjectMembers.ts
│   └── ...
├── pages/            # 페이지 컴포넌트
│   ├── auth/         # 로그인/회원가입
│   ├── projects/     # 프로젝트 관련
│   └── issues/       # 이슈 관련
├── contexts/         # React Context
├── types/            # TypeScript 타입 정의
└── lib/              # 유틸리티 함수
```

## 시작하기

### 1. 의존성 설치

```bash
npm install
```

### 2. 환경 변수 설정

`.env` 파일을 생성하고 다음 내용을 추가:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### 3. 개발 서버 실행

```bash
npm run dev
```

서버가 `http://localhost:5174`에서 실행됩니다.

### 4. 빌드

```bash
npm run build
```

## API 연동

### 인증 토큰

모든 API 요청은 `Authorization: Bearer <token>` 헤더를 포함합니다.
토큰은 로그인 시 localStorage에 저장되며, 자동으로 요청에 추가됩니다.

### API 클라이언트 사용 예시

```typescript
import { projectsApi } from '../api/projects';

// 프로젝트 목록 조회
const projects = await projectsApi.list();

// 프로젝트 생성
const newProject = await projectsApi.create({
  name: 'My Project',
  key: 'PROJ',
  description: 'Project description'
});

// 멤버 추가
await projectsApi.addMember(projectId, {
  user_id: userId,
  role: 'member'
});
```

## React Query 사용

TanStack Query를 사용하여 서버 상태를 관리합니다.

```typescript
import { useProjectMembers } from '../hooks/useProjectMembers';

function MembersPage({ projectId }) {
  const { data: members, isLoading } = useProjectMembers(projectId);

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      {members?.map(member => (
        <div key={member.user_id}>{member.user.username}</div>
      ))}
    </div>
  );
}
```

## 스타일링

Tailwind CSS를 사용하여 스타일링합니다.

```tsx
<button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
  Click me
</button>
```

## 타입 정의

모든 타입은 `src/types/index.ts`에 정의되어 있습니다.

```typescript
export interface ProjectMember {
  project_id: number;
  user_id: number;
  role: ProjectRole;
  user?: User;
  invited_by?: number;
  created_at: string;
}

export type ProjectRole = 'owner' | 'admin' | 'member' | 'viewer';
```

## 권한 시스템

프로젝트 멤버 역할:
- **Owner**: 모든 권한 (프로젝트 삭제 포함)
- **Admin**: 프로젝트 관리 (멤버 관리, 설정 변경)
- **Member**: 이슈 생성/수정, 댓글 작성
- **Viewer**: 읽기 전용

## 개발 가이드

### 새로운 API 추가

1. `src/api/` 폴더에 API 함수 작성
2. `src/hooks/` 폴더에 React Query 훅 작성
3. 페이지/컴포넌트에서 훅 사용

### 새로운 페이지 추가

1. `src/pages/` 폴더에 페이지 컴포넌트 작성
2. `src/App.tsx`에 라우트 추가

```tsx
<Route path="/new-page" element={<NewPage />} />
```

## 최근 업데이트 (2025-11-16)

### Session: 프로젝트 멤버 관리 구현

#### 완료된 작업
1. ✅ `useProjectMembers` 훅 작성
   - `useProjectMembers()` - 멤버 목록 조회
   - `useAddMember()` - 멤버 추가
   - `useUpdateMemberRole()` - 역할 변경
   - `useRemoveMember()` - 멤버 제거

2. ✅ 프로젝트 설정 페이지 개선
   - 탭 기반 UI (라벨, 마일스톤, 멤버)
   - 멤버 목록 표시 (사용자 정보, 역할)
   - 역할 변경 드롭다운
   - 멤버 제거 기능

3. ✅ 권한 기반 UI
   - Owner 역할은 변경 불가
   - 현재 사용자는 자신의 역할 변경 불가
   - 현재 사용자는 자신을 제거 불가

#### 다음 작업
- ⬜ 멤버 추가 모달 구현
- ⬜ 사용자 검색 기능
- ⬜ 초대 링크 생성

## 문제 해결

### Vite 빌드 에러

캐시를 삭제하고 재시작:
```bash
rm -rf node_modules/.vite
npm run dev
```

### API 연결 실패

백엔드 서버가 실행 중인지 확인:
```bash
# 백엔드 서버 확인
curl http://localhost:8080/health
```

## 리소스

- [React Documentation](https://react.dev/)
- [TypeScript Documentation](https://www.typescriptlang.org/)
- [TanStack Query Documentation](https://tanstack.com/query/latest)
- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [Vite Documentation](https://vite.dev/)
