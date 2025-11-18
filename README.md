# Flow - 이슈 트래커

프로젝트 기반 이슈 관리 시스템. Jira/Linear와 유사한 칸반 보드 기능과 세밀한 권한 관리를 제공합니다.

## 주요 기능

### 프로젝트 관리
- 프로젝트 생성/수정/삭제
- 프로젝트별 고유 Key (예: PROJ) 자동 생성
- 프로젝트 멤버 관리 (4단계 권한 시스템)
- 칸반 보드 컬럼 커스터마이징

### 이슈 관리
- 이슈 생성/수정/삭제
- 프로젝트 Key 기반 이슈 번호 (예: PROJ-1, PROJ-2)
- 우선순위 (Low, Medium, High, Critical) 및 상태 관리
- 이슈 담당자 지정 및 마일스톤 연결
- 이슈에 라벨 추가/제거
- 칸반 보드에서 드래그 앤 드롭 이동 (낙관적 잠금)
- 첨부파일 업로드 (MIME 타입 검증, 매직 넘버 검사)

### 협업 기능
- 댓글 시스템 (이슈별 토론)
- **Markdown 지원** (이슈 설명 및 댓글에서 Markdown 포맷 사용 가능, XSS 방지 HTML 새니타이저 적용)
- 이모지 반응 (이슈 및 댓글에 리액션 추가)
- 활동 히스토리 자동 기록
- 알림 시스템 (읽음/안읽음 관리)
- 이메일 알림 (이슈 배정, 댓글 추가 시 자동 발송)
- 통합 검색 (이슈, 프로젝트 전체 검색)

### 권한 시스템
프로젝트별로 4단계 역할 관리:
- **Owner**: 프로젝트 소유자 (모든 권한)
- **Admin**: 프로젝트 관리자 (멤버 관리, 설정 변경)
- **Member**: 일반 멤버 (이슈 생성/수정, 댓글 작성)
- **Viewer**: 읽기 전용 (조회만 가능)

## 기술 스택

### Backend
- **언어**: Go 1.21+
- **프레임워크**: 표준 라이브러리 (net/http)
- **데이터베이스**: PostgreSQL 14+
- **캐시**: Redis 7+
- **인증**: JWT (Access Token + Refresh Token)
- **이메일**: SMTP (선택사항, HTML 템플릿 지원)
- **저장소**: 로컬 파일 시스템

### 아키텍처
```
cmd/server/          # 애플리케이션 진입점
internal/
  ├── api/           # HTTP 라우팅 및 핸들러
  ├── auth/          # JWT 인증
  ├── models/        # 데이터 모델
  ├── repository/    # 데이터베이스 레이어
  └── service/       # 비즈니스 로직
pkg/
  ├── cache/         # Redis 캐시
  ├── database/      # DB 연결 관리
  ├── email/         # SMTP 이메일 클라이언트
  ├── errors/        # 커스텀 에러
  └── storage/       # 파일 저장소
migrations/          # 데이터베이스 마이그레이션
```

## 시작하기

### 1. 필수 요구사항

- Go 1.21 이상
- Docker & Docker Compose (로컬 개발용)
- golang-migrate (마이그레이션용)

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

### 2. 프로젝트 클론

```bash
git clone <repository-url>
cd flow
```

### 3. 환경 설정

```bash
# .env 파일 생성
cp .env.example .env

# .env 파일 수정 (필요시)
# - JWT_SECRET: 운영 환경에서는 반드시 변경
# - DATABASE_URL: 데이터베이스 연결 정보
# - REDIS_URL: Redis 연결 정보
```

### 4. 개발 환경 실행

```bash
# PostgreSQL & Redis 시작
make dev-up

# 데이터베이스 마이그레이션 실행
make migrate-up

# 의존성 설치
go mod download

# 서버 실행
make run

# 또는 빌드 후 실행
make build
./bin/issue-tracker
```

서버가 `http://localhost:8080`에서 실행됩니다.

### 5. 개발 환경 정리

```bash
# 서비스 중지
make dev-down

# 빌드 파일 정리
make clean
```

## API 엔드포인트

### 인증
```
POST   /api/v1/auth/register          # 회원가입
POST   /api/v1/auth/login             # 로그인
POST   /api/v1/auth/refresh           # 토큰 갱신
GET    /api/v1/auth/me                # 내 정보 조회
```

### 프로젝트
```
POST   /api/v1/projects               # 프로젝트 생성
GET    /api/v1/projects               # 프로젝트 목록
GET    /api/v1/projects/{id}          # 프로젝트 조회
PUT    /api/v1/projects/{id}          # 프로젝트 수정
DELETE /api/v1/projects/{id}          # 프로젝트 삭제
```

### 이슈
```
POST   /api/v1/projects/{projectId}/issues     # 이슈 생성
GET    /api/v1/projects/{projectId}/issues     # 이슈 목록
GET    /api/v1/issues/{id}                     # 이슈 조회
GET    /api/v1/issues/{projectKey}/{number}    # Key로 이슈 조회 (예: PROJ-1)
PUT    /api/v1/issues/{id}                     # 이슈 수정
DELETE /api/v1/issues/{id}                     # 이슈 삭제
PUT    /api/v1/issues/{id}/move                # 이슈 보드 이동
```

### 댓글
```
POST   /api/v1/issues/{issueId}/comments       # 댓글 작성
GET    /api/v1/issues/{issueId}/comments       # 댓글 목록
PUT    /api/v1/comments/{id}                   # 댓글 수정
DELETE /api/v1/comments/{id}                   # 댓글 삭제
```

### 라벨
```
POST   /api/v1/projects/{projectId}/labels     # 라벨 생성
GET    /api/v1/projects/{projectId}/labels     # 라벨 목록
PUT    /api/v1/labels/{id}                     # 라벨 수정
DELETE /api/v1/labels/{id}                     # 라벨 삭제
POST   /api/v1/issues/{issueId}/labels/{labelId}    # 라벨 추가
DELETE /api/v1/issues/{issueId}/labels/{labelId}    # 라벨 제거
```

### 첨부파일
```
POST   /api/v1/issues/{id}/attachments         # 파일 업로드
GET    /api/v1/issues/{id}/attachments         # 첨부파일 목록
GET    /api/v1/attachments/{id}/download       # 파일 다운로드
DELETE /api/v1/attachments/{id}                # 파일 삭제
```

### 반응 (Reactions)
```
POST   /api/v1/reactions/{entity_type}/{entity_id}         # 반응 추가/제거 (토글)
GET    /api/v1/reactions/{entity_type}/{entity_id}         # 반응 목록 조회
GET    /api/v1/reactions/{entity_type}/{entity_id}/summary # 반응 요약 조회
DELETE /api/v1/reactions/{entity_type}/{entity_id}/{emoji} # 반응 삭제
```
**지원 이모지**: thumbs_up, thumbs_down, laugh, hooray, confused, heart, rocket, eyes
**entity_type**: issue 또는 comment

### 보드
```
GET    /api/v1/projects/{projectId}/board      # 보드 조회
POST   /api/v1/projects/{projectId}/board/columns   # 컬럼 생성
PUT    /api/v1/board/columns/{id}              # 컬럼 수정
DELETE /api/v1/board/columns/{id}              # 컬럼 삭제
```

### 멤버 관리
```
GET    /api/v1/projects/{projectId}/members    # 멤버 목록
POST   /api/v1/projects/{projectId}/members    # 멤버 추가
PUT    /api/v1/projects/{projectId}/members/{userId}   # 역할 변경
DELETE /api/v1/projects/{projectId}/members/{userId}   # 멤버 제거
```

### 검색
```
GET    /api/v1/search                          # 통합 검색
GET    /api/v1/search/issues                   # 이슈 검색
GET    /api/v1/search/projects                 # 프로젝트 검색
```

### API 문서

Swagger UI를 통해 대화형 API 문서를 확인할 수 있습니다:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`

Swagger 문서 재생성:
```bash
~/go/bin/swag init -g cmd/server/main.go --output docs
```

## 권한 매트릭스

| 기능 | Owner | Admin | Member | Viewer |
|------|-------|-------|--------|--------|
| 프로젝트 조회 | ✅ | ✅ | ✅ | ✅ |
| 프로젝트 수정/삭제 | ✅ | ❌ | ❌ | ❌ |
| 이슈 조회 | ✅ | ✅ | ✅ | ✅ |
| 이슈 생성/수정 | ✅ | ✅ | ✅ | ❌ |
| 이슈 삭제 | ✅ | ✅ | ❌ | ❌ |
| 댓글 조회 | ✅ | ✅ | ✅ | ✅ |
| 댓글 작성/수정 | ✅ | ✅ | ✅ | ❌ |
| 댓글 삭제 (본인) | ✅ | ✅ | ✅ | ❌ |
| 댓글 삭제 (타인) | ✅ | ✅ | ❌ | ❌ |
| 라벨 조회 | ✅ | ✅ | ✅ | ✅ |
| 라벨 관리 | ✅ | ✅ | ❌ | ❌ |
| 첨부파일 조회 | ✅ | ✅ | ✅ | ✅ |
| 첨부파일 업로드 | ✅ | ✅ | ✅ | ❌ |
| 첨부파일 삭제 (본인) | ✅ | ✅ | ✅ | ❌ |
| 첨부파일 삭제 (타인) | ✅ | ✅ | ❌ | ❌ |
| 보드 조회 | ✅ | ✅ | ✅ | ✅ |
| 보드 컬럼 관리 | ✅ | ✅ | ❌ | ❌ |
| 마일스톤 조회 | ✅ | ✅ | ✅ | ✅ |
| 마일스톤 생성/수정 | ✅ | ✅ | ✅ | ❌ |
| 마일스톤 삭제 | ✅ | ✅ | ❌ | ❌ |
| 멤버 관리 | ✅ | ✅ | ❌ | ❌ |

## 데이터베이스 마이그레이션

### 새 마이그레이션 생성
```bash
make migrate-create name=add_new_feature
```

### 마이그레이션 실행
```bash
# 최신 버전으로 업그레이드
make migrate-up

# 한 단계 다운그레이드
make migrate-down
```

## 테스트

```bash
# 전체 테스트 실행
make test

# 커버리지 리포트 생성
make test-coverage
```

## 배포

### 프로덕션 빌드
```bash
make build
```

### 환경 변수 설정
프로덕션 환경에서는 반드시 다음 값들을 변경하세요:

- `JWT_SECRET`: 강력한 랜덤 문자열 (최소 32자)
- `DATABASE_URL`: 실제 데이터베이스 연결 정보
- `REDIS_URL`: 실제 Redis 연결 정보
- `CORS_ALLOWED_ORIGINS`: 허용할 프론트엔드 도메인

#### 이메일 알림 설정 (선택사항)
이메일 알림을 사용하려면 SMTP 서버 정보를 설정하세요:

- `SMTP_HOST`: SMTP 서버 주소 (예: smtp.gmail.com)
- `SMTP_PORT`: SMTP 포트 (기본: 587)
- `SMTP_USERNAME`: SMTP 사용자명
- `SMTP_PASSWORD`: SMTP 비밀번호
- `SMTP_FROM`: 발신자 이메일 주소

**참고**: SMTP 설정이 없어도 시스템은 정상 작동하며, 이메일만 발송되지 않습니다.

### Docker 배포 (선택사항)

```bash
# 프로덕션용 docker-compose 실행
docker-compose -f docker-compose.prod.yml up -d
```

## 보안 기능

- **JWT 인증**: Access Token (15분) + Refresh Token (7일)
- **비밀번호 해싱**: bcrypt 사용
- **API Rate Limiting**: Redis 기반 IP별 요청 제한 (기본: 100 req/분)
  - X-Forwarded-For, X-Real-IP 헤더 지원
  - Rate limit 정보 헤더 반환 (X-RateLimit-Limit, Remaining, Reset)
  - 429 Too Many Requests 응답
  - .env에서 활성화/비활성화 가능
- **파일 업로드 검증**:
  - MIME 타입 화이트리스트
  - 매직 넘버 검사 (파일 확장자 위조 방지)
  - 위험 확장자 차단 (.exe, .sh, .bat 등)
  - 파일 크기 제한 (기본 10MB)
- **SQL Injection 방지**: Prepared Statements 사용
- **CORS 설정**: 허용된 도메인만 접근 가능
- **역할 기반 접근 제어 (RBAC)**: 프로젝트별 세밀한 권한 관리

## 성능 최적화

- **Redis 캐싱**: 프로젝트 목록, 통계, 검색 결과 캐싱
- **데이터베이스 인덱스**: 자주 조회되는 컬럼에 인덱스 설정
- **Connection Pooling**: 데이터베이스 연결 풀 (기본 25개)
- **낙관적 잠금**: 동시성 제어 (이슈 이동 시)
- **N+1 쿼리 방지**: JOIN을 사용한 효율적 쿼리

## 문제 해결

### 데이터베이스 연결 실패
```bash
# PostgreSQL이 실행 중인지 확인
docker-compose ps

# 로그 확인
make dev-logs
```

### 마이그레이션 오류
```bash
# 현재 마이그레이션 상태 확인
migrate -path migrations -database "${DATABASE_URL}" version

# 강제로 특정 버전으로 설정 (주의!)
migrate -path migrations -database "${DATABASE_URL}" force <version>
```

### 포트 충돌
```bash
# 8080 포트를 사용 중인 프로세스 확인
lsof -i :8080

# .env 파일에서 PORT 변경
PORT=3000
```

## 개발 팁

### 코드 린트
```bash
make lint
```

### 개발 도구 설치
```bash
make install-tools
```

### API 테스트
프로젝트 루트에 제공된 테스트 스크립트 사용:
```bash
# 간단한 API 테스트
./simple_test.sh

# 첨부파일 보안 테스트
./test_attachment_security.sh
```
