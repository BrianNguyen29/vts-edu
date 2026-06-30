-- 000010_question_bank_minimal.sql
-- Minimal question bank/version schema with real MCQ content snapshots.

CREATE TABLE IF NOT EXISTS question_banks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    title text NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS questions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    question_bank_id uuid NOT NULL REFERENCES question_banks(id),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS question_versions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id uuid NOT NULL REFERENCES questions(id),
    version int NOT NULL DEFAULT 1,
    prompt_json jsonb NOT NULL,
    choices_json jsonb NOT NULL,
    answer_key_json jsonb NOT NULL,
    max_score numeric(10,2) NOT NULL DEFAULT '1.00',
    status text NOT NULL DEFAULT 'PUBLISHED' CHECK (status IN ('DRAFT','PUBLISHED','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (question_id, version)
);

CREATE TRIGGER trg_question_banks_updated_at
    BEFORE UPDATE ON question_banks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_questions_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DO $$
DECLARE
    org_id uuid;
    bank_id uuid;
    q1_id uuid;
    q2_id uuid;
BEGIN
    SELECT id INTO org_id FROM organizations WHERE code = 'school-a';
    IF org_id IS NULL THEN
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM question_banks
        WHERE organization_id = org_id AND title = 'Bộ câu hỏi Demo Toán'
    ) THEN
        RETURN;
    END IF;

    INSERT INTO question_banks (organization_id, title)
    VALUES (org_id, 'Bộ câu hỏi Demo Toán')
    RETURNING id INTO bank_id;

    INSERT INTO questions (question_bank_id) VALUES (bank_id) RETURNING id INTO q1_id;
    INSERT INTO questions (question_bank_id) VALUES (bank_id) RETURNING id INTO q2_id;

    INSERT INTO question_versions (question_id, version, prompt_json, choices_json, answer_key_json, max_score)
    VALUES (
        q1_id,
        1,
        '{"text":"Giá trị của 5 - 4 bằng bao nhiêu?"}',
        '[{"id":"A","text":"1"},{"id":"B","text":"2"},{"id":"C","text":"3"},{"id":"D","text":"4"}]',
        '{"correct_option":"A"}',
        1.00
    );

    INSERT INTO question_versions (question_id, version, prompt_json, choices_json, answer_key_json, max_score)
    VALUES (
        q2_id,
        1,
        '{"text":"Số nào là số chẵn?"}',
        '[{"id":"A","text":"3"},{"id":"B","text":"4"},{"id":"C","text":"5"},{"id":"D","text":"7"}]',
        '{"correct_option":"B"}',
        1.00
    );
END $$;
