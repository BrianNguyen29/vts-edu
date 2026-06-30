-- 000004_auth_passwords_and_demo_seed.sql
-- Adds password_hash to membership_login_names and seeds demo auth + assessment data.

ALTER TABLE membership_login_names
ADD COLUMN IF NOT EXISTS password_hash text;

DO $$
DECLARE
    org_id uuid;
    user_id uuid;
    membership_id uuid;
    assessment_id uuid;
    publication_id uuid;
    demo_attempt_id uuid;
    existing_login boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM membership_login_names ln
        JOIN organizations o ON o.id = ln.organization_id
        WHERE o.code = 'school-a'
          AND lower(ln.username_normalized) = 'hs001'
    ) INTO existing_login;

    IF existing_login THEN
        RETURN;
    END IF;

    INSERT INTO organizations (code, name)
    VALUES ('school-a', 'Trường THPT Demo A')
    RETURNING id INTO org_id;

    INSERT INTO users DEFAULT VALUES
    RETURNING id INTO user_id;

    INSERT INTO organization_memberships (organization_id, user_id)
    VALUES (org_id, user_id)
    RETURNING id INTO membership_id;

    INSERT INTO membership_login_names (
        organization_id,
        username_normalized,
        user_id,
        password_hash
    ) VALUES (
        org_id,
        'hs001',
        user_id,
        '$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE'
    );

    -- Minimal demo assessment/publication/attempt/items/answers for the exam slice.
    INSERT INTO assessments (organization_id, title, duration_minutes, max_attempts, status)
    VALUES (org_id, 'Bài kiểm tra demo Toán', 45, 1, 'PUBLISHED')
    RETURNING id INTO assessment_id;

    INSERT INTO assessment_publications (organization_id, assessment_id, version, snapshot_json, published_at)
    VALUES (org_id, assessment_id, 1, '{}', now())
    RETURNING id INTO publication_id;

    demo_attempt_id := '00000000-0000-4000-8000-000000000001'::uuid;

    INSERT INTO attempts (id, organization_id, assessment_id, student_user_id, publication_id, status)
    VALUES (demo_attempt_id, org_id, assessment_id, user_id, publication_id, 'CREATED');

    INSERT INTO attempt_items (organization_id, attempt_id, question_version_id, position, points)
    VALUES
        (org_id, demo_attempt_id, gen_random_uuid(), 1, 1.00),
        (org_id, demo_attempt_id, gen_random_uuid(), 2, 1.00);

    INSERT INTO attempt_answers (organization_id, attempt_id, attempt_item_id, answer_payload)
    SELECT org_id, demo_attempt_id, id, '{}'
    FROM attempt_items
    WHERE attempt_id = demo_attempt_id;
END $$;
