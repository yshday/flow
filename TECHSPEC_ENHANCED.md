# Issue Tracker 기술 명세서 (Tech Spec) - Enhanced Version 2.0

**버전**: 2.0 (개선판)  
**작성일**: 2025-11-15  
**기술 스택**: Go + PostgreSQL + Redis + React + TypeScript

> 📌 **주요 개선사항**: 동시성 제어, 권한 관리, JWT Refresh Token, 파일 첨부, 이메일 알림, Full-text Search, 캐싱 전략

---

## 목차

1. [개요 및 개선사항](#1-개요-및-개선사항)
2. [기술 스택 (업데이트)](#2-기술-스택-업데이트)
3. [시스템 아키텍처](#3-시스템-아키텍처)
4. [데이터 모델 (개선)](#4-데이터-모델-개선)
5. [API 설계 (확장)](#5-api-설계-확장)
6. [핵심 구현 가이드](#6-핵심-구현-가이드)
7. [보안 및 권한 관리](#7-보안-및-권한-관리)
8. [성능 최적화](#8-성능-최적화)
9. [모니터링 및 운영](#9-모니터링-및-운영)
10. [개발 우선순위](#10-개발-우선순위)

---

## 1. 개요 및 개선사항

### 1.1 프로젝트 목표
GitHub Issues의 심플함 + Jira의 칸반 보드 + **엔터프라이즈 수준의 안정성**

### 1.2 핵심 개선사항

#### 🔴 심각한 문제 해결
- **동시성 제어**: Race condition 방지를 위한 트랜잭션 처리
- **권한 시스템**: 프로젝트별 역할 기반 접근 제어
- **보안 강화**: JWT Refresh Token, Rate Limiting

#### 🟡 중요 기능 추가
- **파일 첨부**: S3/MinIO 연동
- **이메일 알림**: SMTP 기반 알림 시스템
- **Full-text Search**: PostgreSQL tsvector
- **캐싱**: Redis 기반 다층 캐싱

### 1.3 범위 (v1.0)

✅ **포함 기능**
- 이슈 CRUD with 동시성 제어
- 칸반 보드 with 실시간 동기화
- 프로젝트 권한 관리 (owner/admin/member/viewer)
- JWT with Refresh Token
- 파일 첨부 및 이메일 알림
- Full-text search 및 고급 필터링
- 활동 로그 및 감사

---

## 2. 기술 스택 (업데이트)

### 백엔드 추가 요소
```yaml
캐싱: Redis 7+
파일 스토리지: MinIO/S3
이메일: gomail v2
모니터링: Prometheus
에러 트래킹: Sentry
```

### 프론트엔드 추가 요소
```yaml
에러 트래킹: Sentry
가상 스크롤: react-window
이미지 최적화: sharp
PWA: Workbox
```

---

## 3. 시스템 아키텍처

### 3.1 Clean Architecture 구조

```
internal/
├── domain/                 # 핵심 비즈니스 로직
│   ├── entities/          # 도메인 모델
│   ├── repositories/      # 인터페이스
│   └── services/          # 도메인 서비스
├── usecase/               # 애플리케이션 비즈니스 규칙
├── infrastructure/        # 외부 시스템 구현
│   ├── postgres/         
│   ├── redis/            
│   ├── storage/          
│   └── email/            
└── interfaces/           # 컨트롤러, 프레젠터
    └── http/
        ├── handlers/
        ├── middleware/
        └── routes.go
```

---

## 4. 데이터 모델 (개선)

### 4.1 권한 관리 테이블 (신규)

```sql
-- 프로젝트 멤버 및 권한
CREATE TABLE project_members (
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    joined_at TIMESTAMP DEFAULT NOW(),
    invited_by INTEGER REFERENCES users(id),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_members_user_id ON project_members(user_id);
```

### 4.2 동시성 제어를 위한 이슈 번호 관리 (신규)

```sql
-- 프로젝트별 이슈 카운터 (동시성 안전)
CREATE TABLE project_issue_counters (
    project_id INTEGER PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    last_issue_number INTEGER DEFAULT 0
);

-- 이슈 번호 발급 함수 (ACID 보장)
CREATE OR REPLACE FUNCTION get_next_issue_number(p_project_id INTEGER)
RETURNS INTEGER AS $$
DECLARE
    v_next_number INTEGER;
BEGIN
    UPDATE project_issue_counters
    SET last_issue_number = last_issue_number + 1
    WHERE project_id = p_project_id
    RETURNING last_issue_number INTO v_next_number;
    
    IF NOT FOUND THEN
        INSERT INTO project_issue_counters (project_id, last_issue_number)
        VALUES (p_project_id, 1)
        ON CONFLICT (project_id) 
        DO UPDATE SET last_issue_number = project_issue_counters.last_issue_number + 1
        RETURNING last_issue_number INTO v_next_number;
    END IF;
    
    RETURN v_next_number;
END;
$$ LANGUAGE plpgsql;
```

### 4.3 개선된 이슈 테이블

```sql
CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    issue_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    description_html TEXT, -- XSS 방지 처리된 HTML
    status VARCHAR(20) DEFAULT 'open',
    column_id INTEGER REFERENCES board_columns(id),
    column_position INTEGER, -- 칸반 보드 내 위치
    priority VARCHAR(20) DEFAULT 'medium',
    assignee_id INTEGER REFERENCES users(id),
    reporter_id INTEGER NOT NULL REFERENCES users(id),
    
    -- 추가 필드
    search_vector tsvector, -- Full-text search
    version INTEGER DEFAULT 1, -- 낙관적 락
    estimated_hours DECIMAL(5,2),
    due_date DATE,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP, -- Soft delete
    
    UNIQUE(project_id, issue_number)
);

-- Full-text search 인덱스
CREATE INDEX idx_issues_search ON issues USING GIN(search_vector);

-- Full-text search 자동 업데이트
CREATE TRIGGER update_issues_search_vector
BEFORE INSERT OR UPDATE ON issues
FOR EACH ROW
EXECUTE FUNCTION tsvector_update_trigger(
    search_vector, 'pg_catalog.english', title, description
);
```

### 4.4 파일 첨부 테이블 (신규)

```sql
CREATE TABLE issue_attachments (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    uploaded_by INTEGER NOT NULL REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    storage_path VARCHAR(500) NOT NULL,
    thumbnail_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4.5 JWT Refresh Token 관리 (신규)

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id),
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### 4.6 활동 로그 (개선)

```sql
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id),
    issue_id INTEGER REFERENCES issues(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,
    old_value TEXT,
    new_value TEXT,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 시계열 데이터용 BRIN 인덱스
CREATE INDEX idx_activities_created_at_brin ON activities 
USING BRIN(created_at);
```

---

## 5. API 설계 (확장)

### 5.1 표준 응답 형식

#### 성공 응답
```json
{
  "data": {},
  "meta": {
    "pagination": {
      "cursor": "eyJpZCI6MTAwfQ==",
      "has_more": true,
      "total_count": 150
    }
  }
}
```

#### 에러 응답
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {"field": "email", "message": "Invalid format"}
    ],
    "request_id": "req_1234567890"
  }
}
```

### 5.2 에러 코드 체계

```go
const (
    ErrCodeValidation       = "VALIDATION_ERROR"
    ErrCodeUnauthorized     = "UNAUTHORIZED"
    ErrCodeForbidden        = "FORBIDDEN"
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeConflict         = "CONFLICT"
    ErrCodeRateLimit        = "RATE_LIMIT_EXCEEDED"
    ErrCodeInternal         = "INTERNAL_ERROR"
)
```

### 5.3 주요 엔드포인트

#### 인증 (확장)
- `POST /auth/register` - 회원가입
- `POST /auth/login` - 로그인
- `POST /auth/refresh` - 토큰 갱신 ⭐
- `POST /auth/logout` - 로그아웃 (토큰 무효화) ⭐
- `POST /auth/verify-email` - 이메일 인증 ⭐

#### 프로젝트 권한
- `GET /projects/:id/members` - 멤버 목록 ⭐
- `POST /projects/:id/members` - 멤버 초대 (admin) ⭐
- `PUT /projects/:id/members/:userId` - 역할 변경 (admin) ⭐
- `DELETE /projects/:id/members/:userId` - 멤버 제거 (admin) ⭐

#### 이슈 관리
- `GET /projects/:projectId/issues?cursor=...` - Cursor 페이지네이션 ⭐
- `POST /issues/:id/move` - 칸반 이동 (트랜잭션) ⭐
- `POST /issues/:id/attachments` - 파일 첨부 ⭐

#### 검색
- `GET /search/issues?q=...` - Full-text search ⭐

---

## 6. 핵심 구현 가이드

### 6.1 동시성 안전 칸반 보드 이동

```go
func (s *BoardService) MoveIssue(ctx context.Context, issueID int, targetColumnID int, position int) error {
    return s.db.Transaction(func(tx *sql.Tx) error {
        // 1. 이슈 잠금 (FOR UPDATE)
        var issue Issue
        err := tx.QueryRowContext(ctx, `
            SELECT id, column_id, column_position 
            FROM issues 
            WHERE id = $1 
            FOR UPDATE
        `, issueID).Scan(&issue.ID, &issue.ColumnID, &issue.Position)
        
        if err != nil {
            return err
        }
        
        // 2. 같은 컬럼 내 이동
        if issue.ColumnID == targetColumnID {
            // 위치 재정렬 로직
            if issue.Position < position {
                _, err = tx.ExecContext(ctx, `
                    UPDATE issues 
                    SET column_position = column_position - 1
                    WHERE column_id = $1 
                    AND column_position > $2 
                    AND column_position <= $3
                `, targetColumnID, issue.Position, position)
            } else {
                _, err = tx.ExecContext(ctx, `
                    UPDATE issues 
                    SET column_position = column_position + 1
                    WHERE column_id = $1 
                    AND column_position >= $2 
                    AND column_position < $3
                `, targetColumnID, position, issue.Position)
            }
        } else {
            // 3. 다른 컬럼으로 이동
            // 원본 컬럼 정리
            _, err = tx.ExecContext(ctx, `
                UPDATE issues 
                SET column_position = column_position - 1
                WHERE column_id = $1 AND column_position > $2
            `, issue.ColumnID, issue.Position)
            
            // 대상 컬럼 공간 확보
            _, err = tx.ExecContext(ctx, `
                UPDATE issues 
                SET column_position = column_position + 1
                WHERE column_id = $1 AND column_position >= $2
            `, targetColumnID, position)
        }
        
        // 4. 이슈 위치 업데이트
        _, err = tx.ExecContext(ctx, `
            UPDATE issues 
            SET column_id = $1, column_position = $2, 
                updated_at = NOW(), version = version + 1
            WHERE id = $3
        `, targetColumnID, position, issueID)
        
        // 5. 활동 로그
        _, err = tx.ExecContext(ctx, `
            INSERT INTO activities (issue_id, user_id, action, old_value, new_value)
            VALUES ($1, $2, 'moved', $3, $4)
        `, issueID, userID, issue.ColumnID, targetColumnID)
        
        return err
    })
}
```

### 6.2 낙관적 락 구현

```go
func (s *IssueService) UpdateIssue(ctx context.Context, issue *Issue) error {
    result, err := s.db.ExecContext(ctx, `
        UPDATE issues 
        SET title = $1, description = $2, version = version + 1
        WHERE id = $3 AND version = $4
    `, issue.Title, issue.Description, issue.ID, issue.Version)
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrConcurrentUpdate
    }
    
    return nil
}
```

### 6.3 JWT with Refresh Token

```go
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
}

func (j *JWTManager) GenerateTokenPair(userID int) (*TokenPair, error) {
    // Access Token (15분)
    accessClaims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(15 * time.Minute).Unix(),
        "type":    "access",
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessString, _ := accessToken.SignedString(j.accessSecret)
    
    // Refresh Token (7일)
    refreshID := uuid.New().String()
    refreshClaims := jwt.MapClaims{
        "user_id": userID,
        "jti":     refreshID,
        "exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
        "type":    "refresh",
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshString, _ := refreshToken.SignedString(j.refreshSecret)
    
    // DB에 Refresh Token 저장
    hashedToken := hashToken(refreshString)
    _, err := j.db.Exec(`
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
    `, userID, hashedToken, time.Now().Add(7*24*time.Hour))
    
    return &TokenPair{
        AccessToken:  accessString,
        RefreshToken: refreshString,
        ExpiresIn:    900, // 15분
    }, nil
}
```

### 6.4 권한 미들웨어

```go
func RequireProjectRole(roles ...ProjectRole) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserIDFromContext(r.Context())
            projectID := getProjectIDFromPath(r)
            
            var userRole ProjectRole
            err := db.QueryRow(`
                SELECT role FROM project_members 
                WHERE project_id = $1 AND user_id = $2
            `, projectID, userID).Scan(&userRole)
            
            if err != nil || !hasPermission(userRole, roles) {
                respondError(w, http.StatusForbidden, "Insufficient permissions")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 7. 보안 및 권한 관리

### 7.1 보안 체크리스트

✅ **인증/인가**
- JWT Access Token (15분)
- JWT Refresh Token (7일, Rotation)
- 프로젝트별 역할 기반 접근 제어
- 세션 블랙리스트

✅ **입력 검증**
- SQL Injection 방지 (파라미터화된 쿼리)
- XSS 방지 (HTML Sanitization)
- CSRF 보호
- 파일 업로드 검증 (타입, 크기)

✅ **네트워크**
- HTTPS 강제
- CORS 설정
- Rate Limiting (Redis)
- 보안 헤더 설정

### 7.2 Rate Limiting 구현

```go
func RateLimitMiddleware(redis *redis.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := fmt.Sprintf("rate:%s:%s", getClientIP(r), r.URL.Path)
            
            count, _ := redis.Incr(r.Context(), key).Result()
            if count == 1 {
                redis.Expire(r.Context(), key, time.Minute)
            }
            
            if count > 100 { // 분당 100 요청
                w.Header().Set("X-RateLimit-Limit", "100")
                w.Header().Set("X-RateLimit-Remaining", "0")
                w.Header().Set("Retry-After", "60")
                respondError(w, 429, "Rate limit exceeded")
                return
            }
            
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", 100-count))
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 8. 성능 최적화

### 8.1 캐싱 전략

#### Write-Through Cache
```go
func (c *CacheService) UpdateIssue(ctx context.Context, issue *Issue) error {
    // 1. DB 업데이트
    if err := c.db.UpdateIssue(ctx, issue); err != nil {
        return err
    }
    
    // 2. 캐시 업데이트
    key := fmt.Sprintf("issue:%d", issue.ID)
    c.redis.Set(ctx, key, issue, 1*time.Hour)
    
    // 3. 관련 캐시 무효화
    c.redis.Del(ctx, fmt.Sprintf("project:%d:issues*", issue.ProjectID))
    
    return nil
}
```

### 8.2 Cursor-based 페이지네이션

```go
func (r *IssueRepository) List(ctx context.Context, projectID int, cursor string, limit int) ([]*Issue, string, error) {
    query := `
        SELECT id, title, created_at
        FROM issues
        WHERE project_id = $1
    `
    args := []interface{}{projectID}
    
    if cursor != "" {
        decoded, _ := base64.StdEncoding.DecodeString(cursor)
        query += ` AND (created_at, id) < ($2, $3)`
        // cursor에서 timestamp와 id 추출
    }
    
    query += ` ORDER BY created_at DESC, id DESC LIMIT $4`
    args = append(args, limit+1)
    
    // 쿼리 실행 및 다음 cursor 생성
}
```

### 8.3 데이터베이스 인덱스 전략

```sql
-- 복합 인덱스 (자주 함께 사용되는 필터)
CREATE INDEX idx_issues_project_status_assignee 
ON issues(project_id, status, assignee_id) 
WHERE deleted_at IS NULL;

-- 부분 인덱스 (특정 조건)
CREATE INDEX idx_issues_open 
ON issues(project_id) 
WHERE status = 'open' AND deleted_at IS NULL;

-- BRIN 인덱스 (시계열 데이터)
CREATE INDEX idx_activities_created_at_brin 
ON activities USING BRIN(created_at);
```

---

## 9. 모니터링 및 운영

### 9.1 구조화된 로깅

```go
func NewLogger(env string) *slog.Logger {
    var handler slog.Handler
    
    if env == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, nil)
    }
    
    return slog.New(handler)
}

// 사용 예시
logger.Info("http_request",
    "method", r.Method,
    "path", r.URL.Path,
    "status", status,
    "duration_ms", duration.Milliseconds(),
    "request_id", requestID,
)
```

### 9.2 Prometheus 메트릭

```go
var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "endpoint"},
    )
)
```

### 9.3 Health Check

```go
func HealthCheck(db *sql.DB, redis *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        checks := map[string]string{}
        
        // Database check
        if err := db.Ping(); err != nil {
            checks["database"] = "unhealthy"
        } else {
            checks["database"] = "healthy"
        }
        
        // Redis check
        if err := redis.Ping(r.Context()).Err(); err != nil {
            checks["redis"] = "unhealthy"
        } else {
            checks["redis"] = "healthy"
        }
        
        status := http.StatusOK
        for _, check := range checks {
            if check != "healthy" {
                status = http.StatusServiceUnavailable
                break
            }
        }
        
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(checks)
    }
}
```

---

## 10. 개발 우선순위

### Phase 1: Core MVP (3-4주)

**Week 1: 인프라 & 인증**
- [ ] 프로젝트 구조 (Clean Architecture)
- [ ] DB 스키마 with 동시성 제어
- [ ] JWT with Refresh Token
- [ ] 권한 시스템

**Week 2: 프로젝트 & 이슈**
- [ ] 프로젝트 CRUD with 권한
- [ ] 이슈 CRUD with 동시성 안전 번호 발급
- [ ] 프로젝트 멤버 관리

**Week 3: 칸반 보드**
- [ ] 칸반 컬럼 관리
- [ ] 드래그앤드롭 with 트랜잭션
- [ ] 실시간 동기화 (Polling)

**Week 4: 검색 & 알림**
- [ ] Full-text Search
- [ ] 파일 첨부
- [ ] 이메일 알림
- [ ] 활동 로그

### Phase 2: 안정화 (2주)

**Week 5-6:**
- [ ] Redis 캐싱
- [ ] Rate Limiting
- [ ] 모니터링 (Prometheus)
- [ ] 테스트 (80% coverage)
- [ ] Docker 배포

---

## 부록

### A. 환경 변수

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/issuetracker?sslmode=disable
DB_MAX_CONNECTIONS=25

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_ACCESS_SECRET=your-256-bit-access-secret
JWT_REFRESH_SECRET=your-256-bit-refresh-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d

# Storage
STORAGE_TYPE=s3
S3_BUCKET=issue-tracker-files
S3_REGION=us-east-1

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=

# Security
CORS_ALLOWED_ORIGINS=http://localhost:3000
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

### B. Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: issue_tracker
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://postgres:postgres@postgres:5432/issue_tracker
      REDIS_URL: redis://redis:6379/0
    depends_on:
      - postgres
      - redis
      - minio
    ports:
      - "8080:8080"

  frontend:
    build: ./frontend
    environment:
      VITE_API_URL: http://localhost:8080/api/v1
    ports:
      - "3000:80"

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

### C. 트러블슈팅 가이드

#### 이슈 번호 중복
```sql
-- 카운터 재설정
UPDATE project_issue_counters 
SET last_issue_number = (
    SELECT MAX(issue_number) FROM issues 
    WHERE issues.project_id = project_issue_counters.project_id
);
```

#### 칸반 보드 위치 불일치
```sql
-- 위치 재정렬
WITH numbered AS (
  SELECT id, ROW_NUMBER() OVER (
    PARTITION BY column_id 
    ORDER BY column_position, id
  ) - 1 as new_position
  FROM issues
)
UPDATE issues 
SET column_position = numbered.new_position
FROM numbered
WHERE issues.id = numbered.id;
```

#### 캐시 초기화
```bash
redis-cli FLUSHDB
```

---

## 결론

이 개선된 테크스펙은 실제 프로덕션 환경에서 운영 가능한 견고한 이슈 트래킹 시스템을 구축하기 위한 포괄적인 가이드입니다.

### 핵심 개선 사항
1. **동시성 제어**: Race condition 완벽 방지
2. **권한 시스템**: 엔터프라이즈급 접근 제어
3. **보안 강화**: 다층 보안 체계
4. **성능 최적화**: 캐싱, 인덱싱, 페이지네이션
5. **운영 준비**: 모니터링, 로깅, 백업

**작성자**: 개발팀  
**최종 수정**: 2025-11-15  
**버전**: 2.0 (Enhanced)
