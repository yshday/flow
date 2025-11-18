-- 프로젝트별 이슈 카운터 테이블
CREATE TABLE project_issue_counters (
    project_id INTEGER PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    last_issue_number INTEGER DEFAULT 0
);

-- 이슈 번호 발급 함수 (동시성 안전)
CREATE OR REPLACE FUNCTION get_next_issue_number(p_project_id INTEGER)
RETURNS INTEGER AS $$
DECLARE
    v_next_number INTEGER;
BEGIN
    -- 업데이트하면서 새 번호 받기
    UPDATE project_issue_counters
    SET last_issue_number = last_issue_number + 1
    WHERE project_id = p_project_id
    RETURNING last_issue_number INTO v_next_number;

    -- 레코드가 없으면 새로 생성
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

-- 이슈 생성 시 자동으로 번호 발급하는 트리거
CREATE OR REPLACE FUNCTION set_issue_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.issue_number IS NULL THEN
        NEW.issue_number := get_next_issue_number(NEW.project_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_issue_number
BEFORE INSERT ON issues
FOR EACH ROW
WHEN (NEW.issue_number IS NULL)
EXECUTE FUNCTION set_issue_number();
