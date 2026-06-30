-- 000003_demo_assessments.sql
-- Minimal assessment/attempt/answer/idempotency/audit baseline for the demo scaffold.
-- Full schema will expand in later migrations per backend spec.

CREATE TABLE IF NOT EXISTS assessments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    class_section_id uuid, -- FK will be added when academics schema is introduced
    title text NOT NULL,
    duration_minutes int NOT NULL,
    max_attempts int NOT NULL DEFAULT 1,
    settings_json jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','SCHEDULED','OPEN','CLOSED','GRADING','REVIEWED','PUBLISHED','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS assessment_publications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    assessment_id uuid NOT NULL REFERENCES assessments(id),
    version int NOT NULL,
    snapshot_json jsonb NOT NULL DEFAULT '{}',
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    assessment_id uuid NOT NULL REFERENCES assessments(id),
    student_user_id uuid NOT NULL REFERENCES users(id),
    publication_id uuid REFERENCES assessment_publications(id),
    status text NOT NULL DEFAULT 'CREATED' CHECK (status IN ('CREATED','IN_PROGRESS','SUBMITTED','EXPIRED','TERMINATED')),
    started_at timestamptz,
    expires_at timestamptz,
    submitted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS attempt_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL,
    attempt_id uuid NOT NULL REFERENCES attempts(id),
    question_version_id uuid NOT NULL,
    position int NOT NULL,
    points numeric(10,2) NOT NULL DEFAULT '1.00',
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, attempt_id, position)
);

CREATE TABLE IF NOT EXISTS attempt_answers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL,
    attempt_id uuid NOT NULL REFERENCES attempts(id),
    attempt_item_id uuid NOT NULL REFERENCES attempt_items(id),
    answer_payload jsonb NOT NULL,
    revision bigint NOT NULL DEFAULT 1,
    answered_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, attempt_id, attempt_item_id)
);

CREATE INDEX IF NOT EXISTS idx_attempts_student_assessment
    ON attempts (organization_id, student_user_id, assessment_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_in_progress_expiry
    ON attempts (expires_at)
    WHERE status = 'IN_PROGRESS';

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    actor_user_id uuid NOT NULL REFERENCES users(id),
    scope text NOT NULL,
    key_hash text NOT NULL,
    request_hash text,
    response_status int,
    response_body_json jsonb,
    resource_id uuid,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, actor_user_id, scope, key_hash)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations(id),
    actor_user_id uuid REFERENCES users(id),
    actor_type text,
    action text NOT NULL,
    resource_type text,
    resource_id uuid,
    request_id text,
    ip_prefix_or_hash text,
    user_agent_hash text,
    before_json jsonb,
    after_json jsonb,
    metadata_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_org_action
    ON audit_logs (organization_id, action, created_at DESC);
