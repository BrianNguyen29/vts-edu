-- 000011_attempt_item_content_snapshots.sql
-- Snapshot question prompt/options into attempt_items and wire them to real question_versions.

ALTER TABLE attempt_items
ADD COLUMN IF NOT EXISTS prompt_json jsonb NOT NULL DEFAULT '{}',
ADD COLUMN IF NOT EXISTS choices_json jsonb NOT NULL DEFAULT '{}';

DO $$
DECLARE
    org_id uuid;
    v1_id uuid;
    v2_id uuid;
    item1_id uuid;
    item2_id uuid;
BEGIN
    SELECT id INTO org_id FROM organizations WHERE code = 'school-a';
    IF org_id IS NULL THEN
        RETURN;
    END IF;

    SELECT qv.id INTO v1_id
    FROM question_versions qv
    JOIN questions q ON q.id = qv.question_id
    JOIN question_banks qb ON qb.id = q.question_bank_id
    WHERE qb.organization_id = org_id
      AND qv.answer_key_json ->> 'correct_option' = 'A'
    LIMIT 1;

    SELECT qv.id INTO v2_id
    FROM question_versions qv
    JOIN questions q ON q.id = qv.question_id
    JOIN question_banks qb ON qb.id = q.question_bank_id
    WHERE qb.organization_id = org_id
      AND qv.answer_key_json ->> 'correct_option' = 'B'
    LIMIT 1;

    IF v1_id IS NULL OR v2_id IS NULL THEN
        RETURN;
    END IF;

    SELECT id INTO item1_id
    FROM attempt_items
    WHERE attempt_id = '00000000-0000-4000-8000-000000000001'::uuid
      AND position = 1;

    SELECT id INTO item2_id
    FROM attempt_items
    WHERE attempt_id = '00000000-0000-4000-8000-000000000001'::uuid
      AND position = 2;

    IF item1_id IS NOT NULL THEN
        UPDATE attempt_items
        SET question_version_id = v1_id,
            prompt_json = (SELECT prompt_json FROM question_versions WHERE id = v1_id),
            choices_json = (SELECT choices_json FROM question_versions WHERE id = v1_id),
            answer_key_json = (SELECT answer_key_json FROM question_versions WHERE id = v1_id)
        WHERE id = item1_id;
    END IF;

    IF item2_id IS NOT NULL THEN
        UPDATE attempt_items
        SET question_version_id = v2_id,
            prompt_json = (SELECT prompt_json FROM question_versions WHERE id = v2_id),
            choices_json = (SELECT choices_json FROM question_versions WHERE id = v2_id),
            answer_key_json = (SELECT answer_key_json FROM question_versions WHERE id = v2_id)
        WHERE id = item2_id;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_attempt_items_question_version_id'
          AND table_name = 'attempt_items'
    ) THEN
        ALTER TABLE attempt_items
        ADD CONSTRAINT fk_attempt_items_question_version_id
        FOREIGN KEY (question_version_id) REFERENCES question_versions(id);
    END IF;
END $$;
