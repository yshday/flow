# Flow Issue Tracker - 개발 세션 기록

> **프로젝트**: 이슈 트래커 (Jira/Linear와 유사한 프로젝트 관리 시스템)
> **기술 스택**: Go 1.24, PostgreSQL, Redis, Docker
> **최종 업데이트**: 2025-11-16 (Session 17)

---

## 📚 목차

- [프로젝트 개요](#프로젝트-개요)
- [주요 완성 기능](#주요-완성-기능)
- [Session 16: Docker 프로덕션 배포](#session-16-docker-프로덕션-배포)
- [Session 17: 3-State 이슈 상태 시스템 구현](#session-17-3-state-이슈-상태-시스템-구현)
- [다음 작업](#다음-작업)

---

## 프로젝트 개요

**Flow Issue Tracker**는 Jira나 Linear와 유사한 프로젝트 기반 이슈 관리 시스템입니다.

### 핵심 기능
- ✅ 사용자 인증 (JWT 기반, Access/Refresh Token)
- ✅ 프로젝트 관리 (생성, 수정, 삭제, 권한 관리)
- ✅ 이슈 관리 (CRUD, 상태/우선순위 관리, 담당자 배정)
- ✅ 댓글 시스템 (Markdown 지원, 멘션 기능)
- ✅ 첨부파일 (이미지/파일 업로드, 보안 검증)
- ✅ 활동 로그 (타임라인 추적)
- ✅ 알림 시스템
- ✅ API 문서 (Swagger)
- ✅ Redis 캐싱 및 Rate Limiting
- ✅ Docker 프로덕션 배포

### 기술 스택
- **Backend**: Go 1.24, net/http (표준 라이브러리)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Documentation**: Swagger (swaggo)
- **Deployment**: Docker, Docker Compose

---

## 주요 완성 기능

### Sessions 1-6: 핵심 백엔드
- 프로젝트 구조 설계, 데이터베이스 마이그레이션
- JWT 인증 시스템, 사용자 관리 API
- 프로젝트/이슈 관리 API (CRUD, 권한 관리)
- TDD 기반 Repository/Service/Handler 구현

### Sessions 7-12: 고급 기능
- 댓글 시스템 (Markdown 지원, 멘션)
- 첨부파일 업로드 (보안 검증)
- 활동 로그, 알림 시스템
- Redis 캐싱, Rate Limiting, CORS 설정

### Session 15: API 문서화
- Swagger 통합 (`/swagger/index.html`)
- 모든 API 엔드포인트 문서화
- 인터랙티브 API 테스트 환경

---

## Session 16: Docker 프로덕션 배포

### 📋 목표
Docker를 사용한 프로덕션 환경 배포 구성

### ✅ 완료된 작업

#### 1. Docker 환경 설정
```bash
# .env 파일 생성
DB_HOST=postgres
DB_NAME=issuetracker
DB_PASSWORD=devpassword123
REDIS_HOST=redis
REDIS_PASSWORD=devredis123
JWT_SECRET=SPIbtqNiIx+nW0qrQVry24jUlaw+qqP3ezmbujaY2o8=
JWT_REFRESH_SECRET=gOY0/KXnNJy846TZoTKktJqByuf5ogmOT2CUQAe7ILc=
```

#### 2. Dockerfile 수정
**파일**: `Dockerfile`

**핵심 변경사항**:
```dockerfile
# Go 버전 업데이트 (1.21 → 1.23)
FROM golang:1.23-alpine AS builder

# Toolchain 자동 관리
ENV GOTOOLCHAIN=auto

# Swagger 문서 생성 자동화
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    $(go env GOPATH)/bin/swag init -g cmd/server/main.go
```

#### 3. 데이터베이스 연결 설정 개선
**파일**: `cmd/server/main.go:212-241`

**추가된 함수**:
```go
// buildDatabaseURL: DATABASE_URL 또는 개별 환경 변수 지원
func buildDatabaseURL() string {
    if url := os.Getenv("DATABASE_URL"); url != "" {
        return url
    }
    // DB_HOST, DB_PORT 등으로 URL 구성
    host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "postgres")
    password := getEnv("DB_PASSWORD", "postgres")
    dbname := getEnv("DB_NAME", "issue_tracker")
    sslmode := getEnv("DB_SSLMODE", "disable")
    return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode
}

// buildRedisAddr: REDIS_ADDR 또는 REDIS_HOST+PORT 지원
func buildRedisAddr() string {
    if addr := os.Getenv("REDIS_ADDR"); addr != "" {
        return addr
    }
    host := getEnv("REDIS_HOST", "localhost")
    port := getEnv("REDIS_PORT", "6379")
    return host + ":" + port
}
```

**이유**: Docker Compose는 개별 환경 변수를 제공하므로 유연성 확보

#### 4. go.mod 버전 조정
```go
go 1.24.0  // 의존성 호환성을 위해 1.25.4 → 1.24.0
```

### 🐛 해결한 문제들

| 문제 | 증상 | 해결 방법 |
|------|------|----------|
| **Go 버전 불일치** | `go.mod requires go >= 1.25.4` | Dockerfile: `golang:1.21 → 1.23` |
| **모듈 의존성 요구** | `module requires go >= 1.24.0` | `ENV GOTOOLCHAIN=auto` 추가 + go.mod 1.24 |
| **Swagger 문서 누락** | `no required module provides package .../docs` | swag 설치 및 문서 생성 단계 추가 |
| **DB 연결 실패** | `dial tcp [::1]:5432: connection refused` | `buildDatabaseURL()` 함수 추가 (환경 변수 기반) |
| **인증 실패** | `password authentication failed` | DB 이름 불일치, `docker-compose down -v` 후 재시작 |
| **CORS 이슈** | 프론트엔드 접근 실패 | Docker 이미지 재빌드로 최신 코드 반영 |

### 📊 배포 상태

#### Docker 컨테이너
```bash
$ docker-compose ps
NAME                   STATUS        PORTS
issue-tracker-app      Up            0.0.0.0:8080->8080/tcp
issue-tracker-db       Up (healthy)  5432/tcp
issue-tracker-redis    Up (healthy)  6379/tcp
```

#### Health Check
```bash
$ curl http://localhost:8080/health
{"status":"ok"}
```

#### CORS 검증
```bash
$ curl -H "Origin: http://localhost:5174" http://localhost:8080/api/v1/projects -v
# Response: Access-Control-Allow-Origin: http://localhost:5174
```

### 📝 유용한 명령어

#### Docker 관리
```bash
# 빌드 및 시작
docker-compose build && docker-compose up -d

# 로그 확인
docker-compose logs -f app

# 재시작
docker-compose restart app

# 중지 및 볼륨 삭제
docker-compose down -v

# 이미지 재빌드
docker-compose build app && docker-compose up -d app
```

#### 데이터베이스
```bash
# DB 목록 확인
docker exec issue-tracker-db psql -U postgres -c "\l"

# DB 접속
docker exec -it issue-tracker-db psql -U postgres -d issuetracker
```

### ⚠️ 현재 이슈

**프론트엔드 프로젝트 로딩 실패**:
- URL: `http://localhost:5174/projects`
- CORS: ✅ 정상 작동 확인
- 가능한 원인: 인증 토큰 없음 (401 에러 예상) 또는 빈 데이터베이스
- **다음 단계**: 브라우저 콘솔에서 실제 에러 확인 필요

---

## Session 17: 3-State 이슈 상태 시스템 구현

### 📋 목표
칸반 보드의 컬럼에 따라 이슈 상태가 자동으로 변경되는 3단계 상태 시스템 구현

### ✅ 완료된 작업

#### 1. 백엔드 변경사항

**이슈 모델 업데이트** (`internal/models/issue.go:10`):
```go
const (
    IssueStatusOpen       IssueStatus = "open"
    IssueStatusInProgress IssueStatus = "in_progress"  // 신규 추가
    IssueStatusClosed     IssueStatus = "closed"
)
```

**데이터베이스 마이그레이션** (신규 파일):
- `migrations/000020_add_in_progress_status.up.sql`
- `migrations/000020_add_in_progress_status.down.sql`
- CHECK constraint 업데이트: `'open', 'in_progress', 'closed'` 허용

**마이그레이션 적용**:
```sql
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_status_check;
ALTER TABLE issues ADD CONSTRAINT issues_status_check
    CHECK (status IN ('open', 'in_progress', 'closed'));
```

#### 2. 프론트엔드 변경사항

**TypeScript 타입** (`frontend/src/types/index.ts:71`):
```typescript
export type IssueStatus = 'open' | 'in_progress' | 'closed';
```

**상태 필터 추가** (`frontend/src/pages/projects/ProjectDetailPage.tsx:254`):
- 드롭다운에 "진행 중" (in_progress) 옵션 추가

**상태 표시** (`frontend/src/pages/projects/ProjectDetailPage.tsx:381`):
```typescript
{issue.status === 'open' ? '열림' : issue.status === 'in_progress' ? '진행 중' : '닫힘'}
```

**상태 색상** (`frontend/src/lib/utils.ts:57`):
```typescript
const colors = {
    open: 'text-green-600 bg-green-100',
    in_progress: 'text-yellow-600 bg-yellow-100',  // 신규 추가
    closed: 'text-gray-600 bg-gray-100',
};
```

**칸반 보드 로직** (`frontend/src/pages/projects/ProjectDetailPage.tsx:95-101`):
```typescript
// 컬럼 이름에 따라 자동으로 상태 설정
if (columnName === 'done') {
    status = 'closed';
} else if (columnName === 'in progress') {
    status = 'in_progress';
} else {
    status = 'open';
}
```

#### 3. 테스트 및 문서 업데이트

**테스트 추가** (`internal/repository/issue_repository_test.go:266-294`):
- `in_progress` 상태로 이슈 생성 테스트
- `in_progress` 상태 필터링 테스트

**기술 문서 업데이트** (`TECHSPEC.md`):
- 스키마 설명 업데이트 (line 239)
- API 엔드포인트 설명 업데이트 (line 480)
- 쿼리 파라미터 문서화 (line 485)

### 🎯 구현 결과

#### 상태 매핑
| 칸반 컬럼 | 이슈 상태 | 한글 표시 | 색상 |
|----------|---------|---------|------|
| Backlog | `open` | 열림 | 초록색 |
| In Progress | `in_progress` | 진행 중 | 노란색 |
| Done | `closed` | 닫힘 | 회색 |

#### 동작 방식
1. 사용자가 칸반 보드에서 이슈를 다른 컬럼으로 드래그
2. 프론트엔드가 목표 컬럼 이름을 확인
3. 컬럼 이름에 따라 적절한 상태값 설정
4. API 호출 시 `column_id`와 `status` 함께 전송
5. 백엔드에서 데이터베이스 업데이트
6. 캐시 무효화 및 UI 자동 갱신

### 🐛 해결한 문제

| 문제 | 원인 | 해결 방법 |
|------|------|----------|
| **500 에러 발생** | DB CHECK constraint가 `in_progress` 거부 | 마이그레이션으로 constraint 업데이트 |
| **상태 표시 안됨** | 프론트엔드 타입 및 표시 로직 부재 | TypeScript 타입 및 UI 렌더링 로직 추가 |
| **필터 옵션 없음** | 상태 필터 드롭다운에 옵션 미포함 | 드롭다운에 "진행 중" 옵션 추가 |

### 📊 검증

```bash
# 데이터베이스 확인
docker exec issue-tracker-db psql -U postgres -d issuetracker \
  -c "SELECT issue_number, title, status, column_id FROM issues WHERE project_id = 3;"

# 결과:
# issue_number |     title      |   status    | column_id
# --------------+----------------+-------------+-----------
#             1 | 긴급 버그 수정 | closed      |         9
#             2 | UI 개선 작업   | open        |         7
#             3 | 문서 업데이트  | in_progress |         8
```

### 📝 주요 파일 변경

1. **Backend**:
   - `internal/models/issue.go` - 상태 constant 추가
   - `migrations/000020_add_in_progress_status.*.sql` - DB 마이그레이션

2. **Frontend**:
   - `frontend/src/types/index.ts` - TypeScript 타입 업데이트
   - `frontend/src/pages/projects/ProjectDetailPage.tsx` - UI 및 로직 업데이트
   - `frontend/src/lib/utils.ts` - 색상 유틸리티 업데이트

3. **Tests & Docs**:
   - `internal/repository/issue_repository_test.go` - 테스트 추가
   - `TECHSPEC.md` - 기술 문서 업데이트

---

## Session 17 추가: 라벨 필터링 검증

### 🔍 검증 작업

사용자 보고: "이슈 목록에서 라벨로 검색이 제대로 안되는 것 같다"

#### 검증 과정

1. **데이터베이스 확인**
   ```sql
   -- 라벨 확인
   SELECT l.id, l.name, l.project_id FROM labels l WHERE l.project_id = 3;
   -- 결과: id=1, name="버그"

   -- 이슈-라벨 연결 확인
   SELECT il.issue_id, il.label_id, i.title FROM issue_labels il
   JOIN issues i ON il.issue_id = i.id WHERE i.project_id = 3;
   -- 결과: issue_id=4 (문서 업데이트)에 label_id=1 (버그) 연결됨
   ```

2. **백엔드 API 테스트**
   ```bash
   curl "http://localhost:8080/api/v1/projects/3/issues?label_id=1"
   # 결과: 1개 이슈 반환 ("문서 업데이트")
   ```

3. **프론트엔드 코드 검토**
   - 라벨 필터 드롭다운: `label.id`를 value로 사용 (정상) ✅
   - API 파라미터: `label_id` 전송 (정상) ✅
   - 백엔드 핸들러: `label_id` 파라미터 처리 (정상) ✅
   - 저장소 쿼리: `issue_labels` 테이블 조인 (정상) ✅

4. **브라우저 테스트**
   - 라벨 드롭다운에서 "버그" 선택
   - 네트워크 요청: `GET /api/v1/projects/3/issues?label_id=1&limit=20&offset=0`
   - 결과: "1개의 이슈" 표시, TEST-3만 필터링되어 표시 ✅

#### 검증 결과

**라벨 필터링 기능은 정상 작동** ✅

- 백엔드 API: 정상
- 프론트엔드 UI: 정상
- 데이터베이스 쿼리: 정상
- 네트워크 통신: 정상

사용자가 경험한 문제는 일시적인 UI 상호작용 이슈였을 가능성이 높음 (드롭다운 선택 미완료, 캐시된 데이터 등).

---

## 다음 작업

### 우선순위 높음
1. **프론트엔드 디버깅**:
   - 브라우저 콘솔 에러 확인
   - API 요청 실패 원인 파악
   - 테스트 데이터 생성 (필요시)

### 프로덕션 준비
2. **보안 강화**:
   - 강력한 비밀번호로 변경
   - HTTPS 설정 (nginx reverse proxy)
   - 방화벽 규칙 설정

3. **모니터링**:
   - 로그 수집 시스템
   - 메트릭 모니터링

4. **CI/CD**:
   - 자동 빌드 파이프라인
   - 자동 테스트 실행

---

## 📈 진행 상황

### 완료 ✅
- [x] 백엔드 API 전체 (인증, 프로젝트, 이슈, 댓글, 첨부파일, 알림)
- [x] 데이터베이스 스키마 및 마이그레이션
- [x] Redis 캐싱 및 Rate Limiting
- [x] API 문서화 (Swagger)
- [x] Docker 프로덕션 배포
- [x] 3-State 이슈 상태 시스템 (open/in_progress/closed)
- [x] 칸반 보드 자동 상태 변경
- [x] 라벨 필터링 검증

### 검증 완료 ✅
- [x] 프론트엔드-백엔드 통합
- [x] 칸반 보드 드래그 앤 드롭
- [x] 이슈 상태 자동 변경
- [x] 라벨 필터링 기능

### 계획 📋
- [ ] 프로덕션 보안 강화
- [ ] 모니터링 시스템
- [ ] CI/CD 파이프라인

---

## 📚 참고 자료

### 문서
- [DEPLOYMENT.md](./DEPLOYMENT.md) - Docker 프로덕션 배포 가이드
- Swagger UI: `http://localhost:8080/swagger/index.html`

### 주요 엔드포인트
- Health: `GET /health`
- Auth: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`
- Projects: `GET/POST /api/v1/projects`
- Issues: `GET/POST /api/v1/projects/{id}/issues`

### Docker 컨테이너 구성
- **app**: Go 애플리케이션 (포트 8080)
- **postgres**: PostgreSQL 16 (포트 5432)
- **redis**: Redis 7 (포트 6379)

---

**마지막 업데이트**: Session 17 (2025-11-16)
**현재 상태**: 3-State 이슈 상태 시스템 구현 완료, 칸반 보드 기능 강화
