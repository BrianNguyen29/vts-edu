-- 000020_manual_grading.sql
-- Additive manual-grading foundation for the P1 non-MCQ slice.
-- One grade per attempt item (re-grade is allowed: UPSERT).
-- Scores are stored as numeric(10,2) to match the existing attempts/attempt_items
-- score columns and to avoid float math.

CREATE TABLE IF NOT EXISTS item_grades (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    attempt_id uuid NOT NULL REFERENCES attempts(id) ON DELETE CASCADE,
    attempt_item_id uuid NOT NULL REFERENCES attempt_items(id) ON DELETE CASCADE,
    grader_user_id uuid NOT NULL REFERENCES users(id),
    awarded_score numeric(10, 2) NOT NULL,
    feedback text NULL,
    graded_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT item_grades_awarded_score_nonneg CHECK (awarded_score >= 0),
    CONSTRAINT item_grades_unique_per_item UNIQUE (organization_id, attempt_item_id)
);

CREATE INDEX IF NOT EXISTS item_grades_attempt_idx
    ON item_grades (organization_id, attempt_id, graded_at DESC);

CREATE INDEX IF NOT EXISTS item_grades_grader_idx
    ON item_grades (organization_id, grader_user_id, graded_at DESC);
