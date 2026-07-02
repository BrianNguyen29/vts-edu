-- 000019_question_types.sql
-- Add question_type discriminator to question_versions and snapshot to attempt_items.
-- Backward compatible: existing rows default to 'multiple_choice'.
-- Allow non-MCQ rows to have NULL choices_json / NULL answer_key_json.

ALTER TABLE question_versions
ADD COLUMN IF NOT EXISTS question_type text NOT NULL DEFAULT 'multiple_choice'
CHECK (question_type IN ('multiple_choice', 'short_answer', 'essay'));

ALTER TABLE question_versions
ALTER COLUMN choices_json DROP NOT NULL;

ALTER TABLE question_versions
ALTER COLUMN answer_key_json DROP NOT NULL;

ALTER TABLE attempt_items
ADD COLUMN IF NOT EXISTS question_type text NOT NULL DEFAULT 'multiple_choice'
CHECK (question_type IN ('multiple_choice', 'short_answer', 'essay'));

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'attempts_grading_status_check'
          AND table_name = 'attempts'
    ) THEN
        ALTER TABLE attempts
        ADD CONSTRAINT attempts_grading_status_check
        CHECK (grading_status IS NULL OR grading_status IN ('GRADED', 'PENDING_REVIEW', 'NOT_GRADED'));
    END IF;
END $$;

-- Demo seed: one short_answer and one essay question in the demo bank.
DO $$
DECLARE
    org_id uuid;
    bank_id uuid;
    sa_q_id uuid;
    essay_q_id uuid;
    sa_version_id uuid := '00000000-0000-4000-8000-000000000003'::uuid;
    essay_version_id uuid := '00000000-0000-4000-8000-000000000004'::uuid;
BEGIN
    SELECT id INTO org_id FROM organizations WHERE code = 'school-a';
    IF org_id IS NULL THEN
        RETURN;
    END IF;

    SELECT id INTO bank_id FROM question_banks
    WHERE organization_id = org_id AND title = 'Bộ câu hỏi Demo Toán'
    LIMIT 1;
    IF bank_id IS NULL THEN
        RETURN;
    END IF;

    -- short_answer demo
    IF NOT EXISTS (SELECT 1 FROM question_versions WHERE id = sa_version_id) THEN
        INSERT INTO questions (question_bank_id) VALUES (bank_id) RETURNING id INTO sa_q_id;
        INSERT INTO question_versions (
            id, question_id, version, prompt_json, choices_json, answer_key_json, max_score, status, question_type
        ) VALUES (
            sa_version_id,
            sa_q_id,
            1,
            '{"text":"Viết 3 + 4 bằng mấy? (trả lời ngắn)"}',
            NULL,
            '{"accepted_answers":["7","bảy"]}',
            1.00,
            'PUBLISHED',
            'short_answer'
        );
    END IF;

    -- essay demo
    IF NOT EXISTS (SELECT 1 FROM question_versions WHERE id = essay_version_id) THEN
        INSERT INTO questions (question_bank_id) VALUES (bank_id) RETURNING id INTO essay_q_id;
        INSERT INTO question_versions (
            id, question_id, version, prompt_json, choices_json, answer_key_json, max_score, status, question_type
        ) VALUES (
            essay_version_id,
            essay_q_id,
            1,
            '{"text":"Trình bày cách bạn giải phương trình bậc nhất một ẩn."}',
            NULL,
            '{"rubric":""}',
            2.00,
            'PUBLISHED',
            'essay'
        );
    END IF;
END $$;
