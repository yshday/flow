# Issue Tracker 기술 명세서 (Tech Spec)

**버전**: 1.1
**작성일**: 2025-11-22  
**기술 스택**: Go (표준 라이브러리) + PostgreSQL + React + TypeScript

---

## 목차

1. [개요](#1-개요)
2. [기술 스택](#2-기술-스택)
3. [시스템 아키텍처](#3-시스템-아키텍처)
4. [데이터 모델](#4-데이터-모델)
5. [API 설계](#5-api-설계)
6. [백엔드 구현 가이드](#6-백엔드-구현-가이드)
7. [프론트엔드 구조](#7-프론트엔드-구조)
8. [주요 기능 명세](#8-주요-기능-명세)
9. [개발 우선순위](#9-개발-우선순위)
10. [보안 고려사항](#10-보안-고려사항)

---

## 1. 개요

### 1.1 프로젝트 목표

GitHub Issues 수준의 심플함을 유지하면서 칸반 보드 기반의 시각적 워크플로우를 제공하는 가벼운 이슈 관리 시스템 구축

### 1.2 핵심 원칙

- **심플함**: Jira처럼 무겁지 않고 GitHub Issues처럼 직관적
- **표준 준수**: Go 표준 라이브러리 사용으로 의존성 최소화
- **확장 가능**: 필요시 기능 추가가 용이한 구조
- **타입 안전**: TypeScript와 정적 타입 체킹 활용

### 1.3 범위

#### 포함 기능 (v1.0)
- ✅ 기본 이슈 CRUD
- ✅ 칸반 보드 (드래그 앤 드롭)
- ✅ Milestone, Label 기반 그루핑
- ✅ 코멘트 시스템
- ✅ 검색/필터링
- ✅ 사용자 인증/인가

#### 제외 기능
- ❌ 스프린트 (Milestone로 대체)
- ❌ 고급 리포팅/번다운 차트
- ❌ 타임 트래킹
- ❌ 복잡한 워크플로우 자동화
- ❌ 실시간 협업 (v2 고려)

---

## 2. 기술 스택

### 2.1 백엔드

| 항목 | 기술 | 버전 | 선택 이유 |
|------|------|------|-----------|
| 언어 | Go | 1.25.4+ | 표준 라이브러리 개선, 패턴 매칭 라우팅 |
| HTTP 서버 | net/http | 표준 | 프레임워크 의존성 제거 |
| 데이터베이스 | PostgreSQL | 18.1+ | JSONB, 트랜잭션, 안정성 |
| DB 드라이버 | lib/pq | - | 표준 database/sql 호환 |
| 쿼리 빌더 | sqlc 또는 직접 작성 | - | 타입 안전성, 컴파일 타임 검증 |
| 마이그레이션 | golang-migrate | - | CLI 지원, 롤백 가능 |
| 인증 | JWT (golang-jwt) | v5 | Stateless, 확장 가능 |
| 비밀번호 해싱 | bcrypt | 표준 | 보안 표준 |
| 검증 | validator/v10 | - | 구조체 태그 기반 검증 |
| 로깅 | slog | 표준 | Go 1.21+ 공식 구조화 로깅 |

### 2.2 프론트엔드

| 항목 | 기술 | 버전 | 선택 이유 |
|------|------|------|-----------|
| 프레임워크 | React | 18.3.1+ | 컴포넌트 기반, 생태계 |
| 언어 | TypeScript | 5.9.3+ | 타입 안전성 |
| 빌드 도구 | Vite | 8.1+ | 빠른 개발 서버, HMR |
| 상태 관리 | TanStack Query | v6.0.7+ | 서버 상태 관리, 캐싱 |
| 클라이언트 상태 | Zustand | v5.0.8+ | 가볍고 간단 |
| 라우팅 | Tanstack Router | v1.136+ | 표준적인 선택 |
| UI 라이브러리 | shadcn/ui | v3.5+ | 커스터마이징 가능, Radix UI 기반 |
| 스타일링 | Tailwind CSS | v4.1+ | 유틸리티 퍼스트 |
| 드래그앤드롭 | @dnd-kit | - | 모던하고 접근성 좋음 |
| 폼 관리 | React Hook Form | - | 성능, 검증 통합 |
| 스키마 검증 | Zod | - | TypeScript 퍼스트 |
| HTTP 클라이언트 | Axios | - | 인터셉터, 타입 안전 |

### 2.3 인프라 & DevOps

- **컨테이너**: Docker + Docker Compose
- **API 문서**: OpenAPI 3.0 (수동 또는 생성)
- **테스팅**:
  - Backend: `testing` 패키지, `httptest`
  - Frontend: Vitest, React Testing Library
- **CI/CD**: GitHub Actions (선택)
- **버전 관리**: Git + GitHub

---

## 3. 시스템 아키텍처

### 3.1 전체 구조

```
┌─────────────────┐
│   React SPA     │ ←─── HTTP/REST ───→ ┌──────────────────┐
│   (Frontend)    │                      │   Go HTTP Server │
└─────────────────┘                      │   (Backend)      │
                                         └──────────────────┘
                                                  │
                                                  ↓
                                         ┌──────────────────┐
                                         │   PostgreSQL     │
                                         └──────────────────┘
```

### 3.2 백엔드 레이어 구조

```
cmd/
└── server/
    └── main.go              # 엔트리포인트

internal/
├── api/
│   ├── middleware/          # 인증, 로깅, CORS 등
│   ├── handlers/            # HTTP 핸들러
│   └── routes.go            # 라우팅 설정
├── models/                  # 도메인 모델
├── repository/              # 데이터베이스 접근 계층
├── service/                 # 비즈니스 로직
└── auth/                    # JWT 토큰 생성/검증

pkg/
├── errors/                  # 공통 에러 타입
└── utils/                   # 유틸리티 함수

migrations/                  # 데이터베이스 마이그레이션
```

### 3.3 프론트엔드 구조

```
src/
├── components/              # 재사용 가능한 컴포넌트
│   ├── ui/                  # shadcn/ui 컴포넌트
│   ├── board/               # 칸반 보드 관련
│   ├── issue/               # 이슈 관련
│   └── common/              # 공통 컴포넌트
├── pages/                   # 페이지 컴포넌트
├── hooks/                   # 커스텀 훅
├── api/                     # API 클라이언트
├── stores/                  # Zustand 스토어
├── types/                   # TypeScript 타입 정의
└── lib/                     # 유틸리티
```

---

## 4. 데이터 모델

### 4.1 ERD 개요

```
users ──< projects ──< board_columns
  │                      │
  │                      │
  └──< issues ──< comments
       │  │
       │  └──< issue_labels >──< labels
       │
       └──< milestones
```

### 4.2 테이블 스키마

#### users

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

#### projects

```sql
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    key VARCHAR(10) UNIQUE NOT NULL, -- e.g., "PROJ" for PROJ-1, PROJ-2
    description TEXT,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_projects_key ON projects(key);
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
```

#### board_columns

```sql
CREATE TABLE board_columns (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    position INTEGER NOT NULL, -- 컬럼 순서
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, position)
);

CREATE INDEX idx_board_columns_project_id ON board_columns(project_id);
```

**기본 컬럼**: Backlog (0), In Progress (1), Done (2)

#### issues

```sql
CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_number INTEGER NOT NULL, -- 프로젝트 내 번호 (auto-increment per project)
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open', -- 'open', 'in_progress', 'closed'
    column_id INTEGER REFERENCES board_columns(id),
    priority VARCHAR(20) DEFAULT 'medium', -- 'low', 'medium', 'high', 'urgent'
    issue_type VARCHAR(20) DEFAULT 'task', -- 'bug', 'improvement', 'epic', 'feature', 'task', 'subtask'
    parent_issue_id INTEGER REFERENCES issues(id), -- 서브태스크의 부모 이슈
    epic_id INTEGER REFERENCES issues(id), -- 에픽에 연결된 이슈
    assignee_id INTEGER REFERENCES users(id),
    reporter_id INTEGER NOT NULL REFERENCES users(id),
    milestone_id INTEGER REFERENCES milestones(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, issue_number)
);

CREATE INDEX idx_issues_project_id ON issues(project_id);
CREATE INDEX idx_issues_assignee_id ON issues(assignee_id);
CREATE INDEX idx_issues_status ON issues(status);
CREATE INDEX idx_issues_column_id ON issues(column_id);
CREATE INDEX idx_issues_issue_type ON issues(issue_type);
CREATE INDEX idx_issues_parent_issue_id ON issues(parent_issue_id);
CREATE INDEX idx_issues_epic_id ON issues(epic_id);
```

**이슈 타입 설명**:
| 타입 | 설명 | 아이콘 |
|------|------|--------|
| `task` | 일반 작업 | 📋 |
| `bug` | 결함/버그 | 🐛 |
| `feature` | 신규 기능 | ✨ |
| `improvement` | 기존 기능 개선 | ⚡ |
| `epic` | 대규모 작업 그룹 | 🎯 |
| `subtask` | 하위 작업 | 📎 |

**이슈 번호 자동 증가 트리거**:

```sql
CREATE OR REPLACE FUNCTION set_issue_number()
RETURNS TRIGGER AS $$
BEGIN
    SELECT COALESCE(MAX(issue_number), 0) + 1 INTO NEW.issue_number
    FROM issues WHERE project_id = NEW.project_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_issue_number
BEFORE INSERT ON issues
FOR EACH ROW
WHEN (NEW.issue_number IS NULL)
EXECUTE FUNCTION set_issue_number();
```

#### labels

```sql
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL, -- hex color (e.g., #FF5733)
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_labels_project_id ON labels(project_id);
```

#### issue_labels (Many-to-Many)

```sql
CREATE TABLE issue_labels (
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id INTEGER NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (issue_id, label_id)
);

CREATE INDEX idx_issue_labels_issue_id ON issue_labels(issue_id);
CREATE INDEX idx_issue_labels_label_id ON issue_labels(label_id);
```

#### milestones

```sql
CREATE TABLE milestones (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE,
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'closed'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_milestones_project_id ON milestones(project_id);
CREATE INDEX idx_milestones_status ON milestones(status);
```

#### comments

```sql
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_comments_issue_id ON comments(issue_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
```

#### activities (옵션 - 활동 로그)

```sql
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL, -- 'created', 'status_changed', 'assigned', etc.
    field_name VARCHAR(100), -- 변경된 필드 이름
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_activities_issue_id ON activities(issue_id);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
```

---

## 5. API 설계

### 5.1 API 규칙

- **베이스 URL**: `/api/v1`
- **인증**: `Authorization: Bearer <JWT_TOKEN>` 헤더
- **응답 형식**: JSON
- **에러 형식**:
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Invalid input",
      "details": [...]
    }
  }
  ```

### 5.2 인증 API

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|-----------|
| POST | `/api/v1/auth/register` | 회원가입 | ❌ |
| POST | `/api/v1/auth/login` | 로그인 | ❌ |
| POST | `/api/v1/auth/refresh` | 토큰 갱신 | ✅ |
| GET | `/api/v1/auth/me` | 현재 사용자 정보 | ✅ |

#### 요청/응답 예시

**POST /api/v1/auth/register**

```json
// Request
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "securepass123"
}

// Response (201 Created)
{
  "id": 1,
  "email": "user@example.com",
  "username": "johndoe",
  "created_at": "2025-11-15T10:00:00Z"
}
```

**POST /api/v1/auth/login**

```json
// Request
{
  "email": "user@example.com",
  "password": "securepass123"
}

// Response (200 OK)
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "johndoe"
  }
}
```

### 5.3 프로젝트 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects` | 프로젝트 목록 | ✅ |
| POST | `/api/v1/projects` | 프로젝트 생성 | ✅ |
| GET | `/api/v1/projects/:id` | 프로젝트 상세 | ✅ |
| PUT | `/api/v1/projects/:id` | 프로젝트 수정 | ✅ |
| DELETE | `/api/v1/projects/:id` | 프로젝트 삭제 | ✅ |

**GET /api/v1/projects**

```json
// Response (200 OK)
{
  "projects": [
    {
      "id": 1,
      "name": "Issue Tracker",
      "key": "IT",
      "description": "Main project",
      "owner": {
        "id": 1,
        "username": "johndoe"
      },
      "created_at": "2025-11-01T10:00:00Z"
    }
  ]
}
```

### 5.4 보드 컬럼 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects/:projectId/columns` | 컬럼 목록 | ✅ |
| POST | `/api/v1/projects/:projectId/columns` | 컬럼 생성 | ✅ |
| PUT | `/api/v1/columns/:id` | 컬럼 수정 | ✅ |
| DELETE | `/api/v1/columns/:id` | 컬럼 삭제 | ✅ |
| PATCH | `/api/v1/columns/:id/reorder` | 컬럼 순서 변경 | ✅ |

### 5.5 이슈 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects/:projectId/issues` | 이슈 목록 (필터링, 페이지네이션) | ✅ |
| POST | `/api/v1/projects/:projectId/issues` | 이슈 생성 | ✅ |
| GET | `/api/v1/issues/:id` | 이슈 상세 | ✅ |
| PUT | `/api/v1/issues/:id` | 이슈 수정 | ✅ |
| DELETE | `/api/v1/issues/:id` | 이슈 삭제 | ✅ |
| PATCH | `/api/v1/issues/:id/move` | 이슈 컬럼 이동 | ✅ |
| PATCH | `/api/v1/issues/:id/assign` | 담당자 할당 | ✅ |
| PATCH | `/api/v1/issues/:id/status` | 상태 변경 (open/in_progress/closed) | ✅ |

**GET /api/v1/projects/:projectId/issues**

쿼리 파라미터:
- `status`: open, in_progress, closed, all (기본: open)
- `assignee_id`: 담당자 ID
- `label_ids`: 라벨 ID (쉼표 구분)
- `milestone_id`: 마일스톤 ID
- `priority`: low, medium, high, urgent
- `issue_type`: bug, improvement, epic, feature, task, subtask
- `search`: 제목/설명 검색
- `page`: 페이지 번호 (기본: 1)
- `per_page`: 페이지당 개수 (기본: 20, 최대: 100)

```json
// Response (200 OK)
{
  "issues": [
    {
      "id": 1,
      "project_id": 1,
      "issue_number": 1,
      "title": "Implement user authentication",
      "description": "Add JWT-based auth",
      "status": "open",
      "priority": "high",
      "column": {
        "id": 2,
        "name": "In Progress"
      },
      "assignee": {
        "id": 2,
        "username": "janedoe"
      },
      "reporter": {
        "id": 1,
        "username": "johndoe"
      },
      "labels": [
        {"id": 1, "name": "backend", "color": "#0052CC"}
      ],
      "milestone": {
        "id": 1,
        "title": "v1.0"
      },
      "created_at": "2025-11-10T10:00:00Z",
      "updated_at": "2025-11-12T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

**POST /api/v1/projects/:projectId/issues**

```json
// Request
{
  "title": "Fix login bug",
  "description": "Users cannot login with special characters",
  "priority": "urgent",
  "column_id": 1,
  "assignee_id": 2,
  "label_ids": [1, 3],
  "milestone_id": 1
}

// Response (201 Created)
{
  "id": 46,
  "project_id": 1,
  "issue_number": 46,
  "title": "Fix login bug",
  // ... (전체 이슈 객체)
}
```

**PATCH /api/v1/issues/:id/move**

```json
// Request
{
  "column_id": 3
}

// Response (200 OK)
{
  "id": 1,
  "column": {
    "id": 3,
    "name": "Done"
  },
  "updated_at": "2025-11-15T10:00:00Z"
}
```

### 5.6 에픽 & 서브태스크 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects/:projectId/epics` | 프로젝트 에픽 목록 | ✅ |
| GET | `/api/v1/issues/:epicId/epic-issues` | 에픽에 속한 이슈 목록 | ✅ |
| GET | `/api/v1/issues/:epicId/epic-progress` | 에픽 진행률 | ✅ |
| GET | `/api/v1/issues/:issueId/subtasks` | 서브태스크 목록 | ✅ |
| GET | `/api/v1/issues/:issueId/subtasks/progress` | 서브태스크 진행률 | ✅ |

**GET /api/v1/issues/:epicId/epic-progress**

```json
// Response (200 OK)
{
  "total": 10,
  "completed": 4,
  "in_progress": 3,
  "open": 3,
  "percentage": 40
}
```

### 5.7 라벨 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects/:projectId/labels` | 라벨 목록 | ✅ |
| POST | `/api/v1/projects/:projectId/labels` | 라벨 생성 | ✅ |
| PUT | `/api/v1/labels/:id` | 라벨 수정 | ✅ |
| DELETE | `/api/v1/labels/:id` | 라벨 삭제 | ✅ |

### 5.8 마일스톤 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/projects/:projectId/milestones` | 마일스톤 목록 | ✅ |
| POST | `/api/v1/projects/:projectId/milestones` | 마일스톤 생성 | ✅ |
| GET | `/api/v1/milestones/:id` | 마일스톤 상세 (진행률 포함) | ✅ |
| PUT | `/api/v1/milestones/:id` | 마일스톤 수정 | ✅ |
| DELETE | `/api/v1/milestones/:id` | 마일스톤 삭제 | ✅ |

**GET /api/v1/milestones/:id**

```json
// Response (200 OK)
{
  "id": 1,
  "project_id": 1,
  "title": "v1.0 Release",
  "description": "First stable release",
  "due_date": "2025-12-31",
  "status": "open",
  "progress": {
    "total_issues": 20,
    "closed_issues": 12,
    "percentage": 60
  },
  "created_at": "2025-11-01T10:00:00Z"
}
```

### 5.9 코멘트 API

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/api/v1/issues/:issueId/comments` | 코멘트 목록 | ✅ |
| POST | `/api/v1/issues/:issueId/comments` | 코멘트 작성 | ✅ |
| PUT | `/api/v1/comments/:id` | 코멘트 수정 | ✅ |
| DELETE | `/api/v1/comments/:id` | 코멘트 삭제 | ✅ |

---

## 6. 백엔드 구현 가이드

### 6.1 프로젝트 초기화

```bash
# 프로젝트 생성
mkdir issue-tracker-backend
cd issue-tracker-backend
go mod init github.com/yourusername/issue-tracker

# 필요 패키지 설치
go get github.com/lib/pq
go get github.com/golang-jwt/jwt/v5
go get github.com/go-playground/validator/v10
go get golang.org/x/crypto/bcrypt
```

### 6.2 디렉토리 구조

```
issue-tracker-backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── project.go
│   │   │   ├── issue.go
│   │   │   ├── label.go
│   │   │   ├── milestone.go
│   │   │   └── comment.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logging.go
│   │   └── routes.go
│   ├── models/
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── issue.go
│   │   └── ...
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── project_repo.go
│   │   ├── issue_repo.go
│   │   └── ...
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── project_service.go
│   │   └── issue_service.go
│   └── auth/
│       └── jwt.go
├── pkg/
│   ├── database/
│   │   └── postgres.go
│   ├── errors/
│   │   └── errors.go
│   └── validator/
│       └── validator.go
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   └── ...
├── .env.example
├── docker-compose.yml
└── go.mod
```

### 6.3 주요 코드 예시

#### main.go

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/yourusername/issue-tracker/internal/api"
    "github.com/yourusername/issue-tracker/pkg/database"
)

func main() {
    // 로거 설정
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    // 데이터베이스 연결
    db, err := database.NewPostgresDB(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // 라우터 설정
    mux := api.NewRouter(db)

    // 서버 설정
    srv := &http.Server{
        Addr:         ":" + getEnv("PORT", "8080"),
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    go func() {
        slog.Info("Starting server", "port", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed:", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    slog.Info("Server exited")
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

#### routes.go

```go
package api

import (
    "database/sql"
    "net/http"

    "github.com/yourusername/issue-tracker/internal/api/handlers"
    "github.com/yourusername/issue-tracker/internal/api/middleware"
)

func NewRouter(db *sql.DB) http.Handler {
    mux := http.NewServeMux()

    // 핸들러 초기화
    authHandler := handlers.NewAuthHandler(db)
    projectHandler := handlers.NewProjectHandler(db)
    issueHandler := handlers.NewIssueHandler(db)

    // 공개 라우트
    mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
    mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

    // 보호된 라우트 - 인증 필요
    apiMux := http.NewServeMux()
    
    // 프로젝트
    apiMux.HandleFunc("GET /api/v1/projects", projectHandler.List)
    apiMux.HandleFunc("POST /api/v1/projects", projectHandler.Create)
    apiMux.HandleFunc("GET /api/v1/projects/{id}", projectHandler.Get)
    apiMux.HandleFunc("PUT /api/v1/projects/{id}", projectHandler.Update)
    apiMux.HandleFunc("DELETE /api/v1/projects/{id}", projectHandler.Delete)

    // 이슈
    apiMux.HandleFunc("GET /api/v1/projects/{projectId}/issues", issueHandler.List)
    apiMux.HandleFunc("POST /api/v1/projects/{projectId}/issues", issueHandler.Create)
    apiMux.HandleFunc("GET /api/v1/issues/{id}", issueHandler.Get)
    apiMux.HandleFunc("PUT /api/v1/issues/{id}", issueHandler.Update)
    apiMux.HandleFunc("DELETE /api/v1/issues/{id}", issueHandler.Delete)
    apiMux.HandleFunc("PATCH /api/v1/issues/{id}/move", issueHandler.Move)

    // 미들웨어 체인
    handler := middleware.Chain(
        apiMux,
        middleware.Logging,
        middleware.CORS,
        middleware.Authenticate(db),
    )

    mux.Handle("/api/v1/", handler)

    return mux
}
```

#### middleware/auth.go

```go
package middleware

import (
    "context"
    "database/sql"
    "net/http"
    "strings"

    "github.com/yourusername/issue-tracker/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func Authenticate(db *sql.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Authorization 헤더에서 토큰 추출
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
                return
            }

            token := parts[1]

            // JWT 검증
            claims, err := auth.ValidateToken(token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // 사용자 ID를 컨텍스트에 저장
            ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 컨텍스트에서 사용자 ID 추출
func GetUserID(ctx context.Context) (int, bool) {
    userID, ok := ctx.Value(UserContextKey).(int)
    return userID, ok
}
```

#### middleware/logging.go

```go
package middleware

import (
    "log/slog"
    "net/http"
    "time"
)

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Response writer wrapper to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        slog.Info("HTTP request",
            "method", r.Method,
            "path", r.URL.Path,
            "status", wrapped.statusCode,
            "duration", time.Since(start),
            "remote_addr", r.RemoteAddr,
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

#### middleware/cors.go

```go
package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*") // 프로덕션에서는 특정 도메인으로 제한
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

#### middleware/chain.go

```go
package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}
```

#### auth/jwt.go

```go
package auth

import (
    "errors"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID int `json:"user_id"`
    jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(userID int) (string, error) {
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
```

#### handlers/auth.go (예시)

```go
package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "golang.org/x/crypto/bcrypt"

    "github.com/yourusername/issue-tracker/internal/auth"
    "github.com/yourusername/issue-tracker/internal/models"
)

type AuthHandler struct {
    db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
    return &AuthHandler{db: db}
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // 비밀번호 해싱
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // 사용자 생성
    var user models.User
    err = h.db.QueryRow(`
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, email, username, created_at
    `, req.Email, req.Username, string(hashedPassword)).Scan(
        &user.ID, &user.Email, &user.Username, &user.CreatedAt,
    )

    if err != nil {
        http.Error(w, "User already exists", http.StatusConflict)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    AccessToken string      `json:"access_token"`
    TokenType   string      `json:"token_type"`
    ExpiresIn   int         `json:"expires_in"`
    User        models.User `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // 사용자 조회
    var user models.User
    var passwordHash string
    err := h.db.QueryRow(`
        SELECT id, email, username, password_hash, created_at
        FROM users WHERE email = $1
    `, req.Email).Scan(&user.ID, &user.Email, &user.Username, &passwordHash, &user.CreatedAt)

    if err == sql.ErrNoRows {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 비밀번호 검증
    if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // JWT 생성
    token, err := auth.GenerateToken(user.ID)
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    response := LoginResponse{
        AccessToken: token,
        TokenType:   "Bearer",
        ExpiresIn:   86400, // 24 hours
        User:        user,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### 6.4 환경 변수 (.env)

```env
# Server
PORT=8080

# Database
DATABASE_URL=postgres://user:password@localhost:5432/issue_tracker?sslmode=disable

# JWT
JWT_SECRET=your-secret-key-change-this-in-production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

### 6.5 Docker Compose

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: issue_tracker
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://user:password@db:5432/issue_tracker?sslmode=disable
      JWT_SECRET: your-secret-key
    depends_on:
      - db

volumes:
  postgres_data:
```

---

## 7. 프론트엔드 구조

### 7.1 프로젝트 초기화

```bash
npm create vite@latest issue-tracker-frontend -- --template react-ts
cd issue-tracker-frontend
npm install

# 의존성 설치
npm install react-router-dom
npm install @tanstack/react-query
npm install zustand
npm install axios
npm install react-hook-form zod @hookform/resolvers
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities

# shadcn/ui 설정
npx shadcn-ui@latest init
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add input
npx shadcn-ui@latest add label
npx shadcn-ui@latest add textarea
# 필요한 컴포넌트 추가...
```

### 7.2 폴더 구조

```
src/
├── components/
│   ├── ui/                    # shadcn/ui 컴포넌트
│   ├── layout/
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   └── Layout.tsx
│   ├── board/
│   │   ├── KanbanBoard.tsx
│   │   ├── BoardColumn.tsx
│   │   └── IssueCard.tsx
│   ├── issue/
│   │   ├── IssueList.tsx
│   │   ├── IssueDetail.tsx
│   │   ├── IssueForm.tsx
│   │   └── IssueFilters.tsx
│   ├── comment/
│   │   ├── CommentList.tsx
│   │   └── CommentForm.tsx
│   └── common/
│       ├── LoadingSpinner.tsx
│       ├── ErrorMessage.tsx
│       └── ConfirmDialog.tsx
├── pages/
│   ├── auth/
│   │   ├── LoginPage.tsx
│   │   └── RegisterPage.tsx
│   ├── projects/
│   │   ├── ProjectListPage.tsx
│   │   ├── ProjectDetailPage.tsx
│   │   └── ProjectBoardPage.tsx
│   └── issues/
│       ├── IssueListPage.tsx
│       └── IssueDetailPage.tsx
├── api/
│   ├── client.ts             # Axios 인스턴스
│   ├── auth.ts
│   ├── projects.ts
│   ├── issues.ts
│   └── types.ts              # API 타입 정의
├── hooks/
│   ├── useAuth.ts
│   ├── useProjects.ts
│   └── useIssues.ts
├── stores/
│   └── authStore.ts          # Zustand 스토어
├── lib/
│   └── utils.ts
├── types/
│   └── index.ts
├── App.tsx
└── main.tsx
```

### 7.3 주요 코드 예시

#### api/client.ts

```typescript
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 요청 인터셉터 - JWT 토큰 추가
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 응답 인터셉터 - 401 에러 처리
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

#### stores/authStore.ts

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: number;
  email: string;
  username: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      setAuth: (user, token) => {
        localStorage.setItem('access_token', token);
        set({ user, token });
      },
      logout: () => {
        localStorage.removeItem('access_token');
        set({ user: null, token: null });
      },
      isAuthenticated: () => !!get().token,
    }),
    {
      name: 'auth-storage',
    }
  )
);
```

#### hooks/useIssues.ts

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { issuesApi } from '@/api/issues';
import type { Issue, CreateIssueData } from '@/types';

export function useIssues(projectId: number, filters?: Record<string, any>) {
  return useQuery({
    queryKey: ['issues', projectId, filters],
    queryFn: () => issuesApi.list(projectId, filters),
  });
}

export function useIssue(issueId: number) {
  return useQuery({
    queryKey: ['issues', issueId],
    queryFn: () => issuesApi.get(issueId),
  });
}

export function useCreateIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateIssueData) => issuesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
    },
  });
}

export function useMoveIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ issueId, columnId }: { issueId: number; columnId: number }) =>
      issuesApi.move(issueId, columnId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['issues'] });
    },
  });
}
```

---

## 8. 주요 기능 명세

### 8.1 칸반 보드

**요구사항**:
- 드래그 앤 드롭으로 이슈를 컬럼 간 이동
- 컬럼별 이슈 개수 표시
- 컬럼 순서 변경 가능
- 컬럼 추가/수정/삭제 가능

**기본 컬럼**:
1. Backlog (position: 0)
2. In Progress (position: 1)
3. Done (position: 2)

**구현 고려사항**:
- `@dnd-kit` 라이브러리 사용
- 낙관적 업데이트 (Optimistic Update)
- 드래그 중 시각적 피드백

### 8.2 이슈 상세

**포함 정보**:
- 제목, 설명 (Markdown 렌더링)
- 상태 (open/closed)
- 우선순위
- 담당자, 리포터
- 라벨 (다중)
- 마일스톤
- 생성/수정 날짜
- 코멘트 스레드
- 활동 로그 (옵션)

**액션**:
- 상태 변경
- 담당자 할당/해제
- 라벨 추가/제거
- 마일스톤 설정
- 이슈 삭제

### 8.3 검색 & 필터

**필터 옵션**:
- 상태: open, closed, all
- 담당자: 사용자 선택
- 라벨: 다중 선택 (OR 조건)
- 마일스톤: 단일 선택
- 우선순위: low, medium, high, urgent

**검색**:
- 제목과 설명에서 전체 텍스트 검색
- PostgreSQL `ILIKE` 또는 Full-Text Search 사용

**정렬**:
- 생성일 (최신순/오래된순)
- 수정일
- 우선순위
- 이슈 번호

### 8.4 마일스톤 진행률

**계산 방식**:
```
진행률 = (완료된 이슈 수 / 전체 이슈 수) * 100
```

**표시**:
- 프로그레스 바
- 퍼센티지
- 완료/전체 이슈 수

---

## 9. 개발 우선순위

### Phase 1: MVP (4-6주)

**Week 1-2: 백엔드 기초**
- [ ] 프로젝트 구조 설정
- [ ] 데이터베이스 마이그레이션
- [ ] 사용자 인증 (회원가입, 로그인)
- [ ] 프로젝트 CRUD

**Week 3-4: 이슈 시스템**
- [ ] 이슈 CRUD
- [ ] 기본 칸반 보드 (고정 컬럼)
- [ ] 라벨 시스템
- [ ] 이슈 필터링

**Week 5-6: 프론트엔드**
- [ ] 레이아웃 & 라우팅
- [ ] 로그인/회원가입 페이지
- [ ] 프로젝트 목록/상세 페이지
- [ ] 칸반 보드 페이지 (드래그 앤 드롭)
- [ ] 이슈 리스트 & 상세 페이지

### Phase 2: 고급 기능 (3-4주)

**Week 7-8**
- [ ] 컬럼 커스터마이징
- [ ] 마일스톤 시스템
- [ ] 검색 고도화
- [ ] 코멘트 시스템

**Week 9-10**
- [ ] 활동 로그
- [ ] UI/UX 개선
- [ ] 에러 핸들링 강화
- [ ] 테스트 작성

### Phase 3: 최적화 & 배포 (2주)

**Week 11-12**
- [ ] 성능 최적화
- [ ] Docker 이미지 빌드
- [ ] 배포 (선택: AWS, GCP, Vercel 등)
- [ ] 문서화

---

## 10. 보안 고려사항

### 10.1 인증 & 인가

- **JWT 토큰**:
  - 만료 시간: 24시간
  - Refresh Token 고려 (Phase 2)
  - Secret Key는 환경변수로 관리

- **비밀번호**:
  - bcrypt로 해싱 (cost: 10-12)
  - 최소 8자 이상 요구
  - 특수문자 포함 권장 (선택)

### 10.2 입력 검증

- **백엔드**: validator 라이브러리로 구조체 검증
- **프론트엔드**: Zod 스키마 검증
- **SQL Injection 방지**: 파라미터화된 쿼리 (`$1`, `$2` 등)
- **XSS 방지**: 
  - Markdown 렌더링 시 sanitize
  - React의 기본 이스케이핑 활용

### 10.3 CORS

- 프로덕션: 특정 도메인만 허용
- 개발: `localhost:5173` 허용

### 10.4 HTTPS

- 프로덕션 환경에서는 HTTPS 필수
- Let's Encrypt 인증서 사용 권장

### 10.5 Rate Limiting (Phase 2)

- API 엔드포인트별 요청 제한
- IP 기반 또는 사용자 기반

---

## 11. 향후 확장 가능성

### v2.0 고려 기능

- [ ] 실시간 협업 (WebSocket)
- [ ] 알림 시스템
- [ ] 파일 첨부
- [ ] 이슈 템플릿
- [ ] 커스텀 필드
- [ ] 웹훅 (GitHub, Slack 연동)
- [ ] 다크 모드
- [ ] 모바일 앱 (Flutter)
- [ ] 고급 리포팅 (번다운 차트, 벨로시티)
- [ ] 스프린트 기능

---

## 부록

### A. 환경 변수 전체 목록

```env
# Server
PORT=8080
ENV=development # development, production

# Database
DATABASE_URL=postgres://user:password@localhost:5432/issue_tracker?sslmode=disable

# JWT
JWT_SECRET=your-256-bit-secret
JWT_EXPIRY=24h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Frontend (Vite)
VITE_API_URL=http://localhost:8080/api/v1
```

### B. 유용한 명령어

```bash
# 마이그레이션
migrate -path migrations -database "postgres://..." up
migrate -path migrations -database "postgres://..." down

# 개발 서버 실행
go run cmd/server/main.go

# 프론트엔드 개발 서버
npm run dev

# Docker Compose
docker-compose up -d
docker-compose logs -f backend
```

### C. 참고 자료

- Go 표준 라이브러리: https://pkg.go.dev/std
- PostgreSQL 문서: https://www.postgresql.org/docs/
- React 공식 문서: https://react.dev/
- TanStack Query: https://tanstack.com/query/latest
- shadcn/ui: https://ui.shadcn.com/

---

**작성자**: 개발팀
**최종 수정**: 2025-11-22
**버전**: 1.1
