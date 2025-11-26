-- Project Templates: 프로젝트 생성 시 사용할 수 있는 템플릿
CREATE TABLE project_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,  -- true: 시스템 기본 템플릿, false: 사용자 정의
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,

    -- 템플릿 설정 (JSON)
    -- columns: [{name, position, wip_limit}]
    -- labels: [{name, color, description}]
    -- milestones: [{title, description}]
    config JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Issue Templates: 이슈 생성 시 사용할 수 있는 템플릿
CREATE TABLE issue_templates (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- 템플릿 내용 (마크다운)
    content TEXT NOT NULL DEFAULT '',

    -- 기본 값
    default_priority VARCHAR(20) DEFAULT 'medium',
    default_labels INTEGER[] DEFAULT '{}',  -- 자동 적용될 라벨 ID 배열

    -- 정렬 순서
    position INTEGER DEFAULT 0,

    -- 활성화 여부
    is_active BOOLEAN DEFAULT true,

    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 인덱스
CREATE INDEX idx_project_templates_is_system ON project_templates(is_system);
CREATE INDEX idx_issue_templates_project_id ON issue_templates(project_id);
CREATE INDEX idx_issue_templates_is_active ON issue_templates(is_active);

-- 시스템 기본 프로젝트 템플릿 추가
INSERT INTO project_templates (name, description, is_system, config) VALUES
(
    '칸반 기본',
    '기본적인 칸반 보드 구성입니다. Backlog, In Progress, Done 3개의 컬럼으로 구성됩니다.',
    true,
    '{
        "columns": [
            {"name": "Backlog", "position": 0},
            {"name": "In Progress", "position": 1},
            {"name": "Done", "position": 2}
        ],
        "labels": [
            {"name": "버그", "color": "#dc2626", "description": "버그 수정이 필요한 이슈"},
            {"name": "기능", "color": "#16a34a", "description": "새로운 기능 개발"},
            {"name": "개선", "color": "#2563eb", "description": "기존 기능 개선"}
        ]
    }'::jsonb
),
(
    '스크럼 스프린트',
    '스크럼 방법론을 위한 보드입니다. Sprint Backlog, In Progress, Review, Done으로 구성됩니다.',
    true,
    '{
        "columns": [
            {"name": "Backlog", "position": 0},
            {"name": "Sprint Backlog", "position": 1},
            {"name": "In Progress", "position": 2},
            {"name": "Review", "position": 3},
            {"name": "Done", "position": 4}
        ],
        "labels": [
            {"name": "스토리", "color": "#16a34a", "description": "사용자 스토리"},
            {"name": "태스크", "color": "#2563eb", "description": "개발 태스크"},
            {"name": "버그", "color": "#dc2626", "description": "버그 수정"},
            {"name": "스파이크", "color": "#9333ea", "description": "조사/연구 작업"}
        ]
    }'::jsonb
),
(
    '버그 트래킹',
    'QA 및 버그 관리를 위한 보드입니다. 버그의 라이프사이클을 추적합니다.',
    true,
    '{
        "columns": [
            {"name": "New", "position": 0},
            {"name": "Triaging", "position": 1},
            {"name": "In Progress", "position": 2},
            {"name": "Testing", "position": 3},
            {"name": "Closed", "position": 4}
        ],
        "labels": [
            {"name": "critical", "color": "#dc2626", "description": "치명적인 버그"},
            {"name": "major", "color": "#ea580c", "description": "주요 버그"},
            {"name": "minor", "color": "#ca8a04", "description": "사소한 버그"},
            {"name": "trivial", "color": "#65a30d", "description": "미미한 버그"}
        ]
    }'::jsonb
),
(
    '기능 개발',
    '새로운 기능 개발 프로세스를 위한 보드입니다.',
    true,
    '{
        "columns": [
            {"name": "Idea", "position": 0},
            {"name": "Planning", "position": 1},
            {"name": "Development", "position": 2},
            {"name": "QA", "position": 3},
            {"name": "Release", "position": 4}
        ],
        "labels": [
            {"name": "MVP", "color": "#dc2626", "description": "MVP 필수 기능"},
            {"name": "P1", "color": "#ea580c", "description": "우선순위 높음"},
            {"name": "P2", "color": "#ca8a04", "description": "우선순위 중간"},
            {"name": "P3", "color": "#65a30d", "description": "우선순위 낮음"},
            {"name": "nice-to-have", "color": "#2563eb", "description": "있으면 좋은 기능"}
        ]
    }'::jsonb
);
