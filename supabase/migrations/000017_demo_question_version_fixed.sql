-- 000017_demo_question_version_fixed.sql
-- Adds a deterministic question version for E2E assessment-builder smoke tests.

DO $$
DECLARE
    org_id uuid;
    bank_id uuid;
    q_id uuid;
    fixed_version_id uuid := '00000000-0000-4000-8000-000000000002'::uuid;
BEGIN
    SELECT id INTO org_id FROM organizations WHERE code = 'school-a';
    IF org_id IS NULL THEN
        RETURN;
    END IF;

    SELECT id INTO bank_id FROM question_banks WHERE organization_id = org_id AND title = 'Bộ câu hỏi Demo Toán' LIMIT 1;
    IF bank_id IS NULL THEN
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM question_versions WHERE id = fixed_version_id) THEN
        RETURN;
    END IF;

    INSERT INTO questions (id, question_bank_id)
    VALUES (gen_random_uuid(), bank_id)
    RETURNING id INTO q_id;

    INSERT INTO question_versions (id, question_id, version, prompt_json, choices_json, answer_key_json, max_score)
    VALUES (
        fixed_version_id,
        q_id,
        1,
        '{"text":"Giá trị của 2 + 3 bằng bao nhiêu?"}',
        '[{"id":"A","text":"4"},{"id":"B","text":"5"},{"id":"C","text":"6"},{"id":"D","text":"7"}]',
        '{"correct_option":"B"}',
        1.00
    );
END $$;
