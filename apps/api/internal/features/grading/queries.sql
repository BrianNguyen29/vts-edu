-- name: ListReviewQueue :many
-- Pending-review attempts + the count of ungraded non-MCQ items for each.
SELECT
    a.id AS attempt_id,
    a.organization_id,
    a.assessment_id,
    a.student_user_id,
    u.display_name AS student_name,
    a.status,
    a.started_at,
    a.submitted_at,
    a.expires_at,
    a.max_score,
    COALESCE(pending.pending_items, 0)::int AS pending_items,
    COALESCE(pending.total_non_mcq, 0)::int AS total_non_mcq
FROM attempts a
JOIN users u ON u.id = a.student_user_id
JOIN LATERAL (
    SELECT
        COUNT(*) FILTER (
            WHERE ai.question_type IN ('essay', 'short_answer') AND ig.id IS NULL
        ) AS pending_items,
        COUNT(*) FILTER (
            WHERE ai.question_type IN ('essay', 'short_answer')
        ) AS total_non_mcq
    FROM attempt_items ai
    LEFT JOIN item_grades ig
        ON ig.attempt_item_id = ai.id
        AND ig.organization_id = ai.organization_id
    WHERE ai.attempt_id = a.id
      AND ai.organization_id = a.organization_id
) pending ON true
WHERE a.organization_id = $1
  AND a.assessment_id = $2
  AND a.status IN ('SUBMITTED', 'EXPIRED')
  AND a.grading_status = 'PENDING_REVIEW'
ORDER BY a.submitted_at ASC NULLS LAST, a.started_at ASC;

-- name: GetAttemptForGrading :one
-- Org-scoped attempt (no student_user_id filter) for teacher/admin review.
SELECT
    a.id,
    a.organization_id,
    a.assessment_id,
    a.student_user_id,
    a.publication_id,
    a.status,
    a.started_at,
    a.expires_at,
    a.submitted_at,
    a.score,
    a.max_score,
    a.grading_status
FROM attempts a
WHERE a.id = $1
  AND a.organization_id = $2
LIMIT 1;

-- name: GetAttemptItemsForGrading :many
-- Items in an attempt joined with any current manual grade.
SELECT
    ai.id,
    ai.organization_id,
    ai.attempt_id,
    ai.question_version_id,
    ai.position,
    ai.points::text AS points,
    ai.prompt_json,
    ai.choices_json,
    ai.answer_key_json,
    ai.question_type,
    aa.answer_payload,
    aa.revision,
    aa.answered_at,
    ig.id AS item_grade_id,
    COALESCE(ig.awarded_score, '0')::text AS awarded_score,
    COALESCE(ig.feedback, '') AS feedback,
    ig.grader_user_id,
    ig.graded_at
FROM attempt_items ai
LEFT JOIN attempt_answers aa
    ON aa.attempt_item_id = ai.id
    AND aa.organization_id = ai.organization_id
    AND aa.attempt_id = ai.attempt_id
LEFT JOIN item_grades ig
    ON ig.attempt_item_id = ai.id
    AND ig.organization_id = ai.organization_id
WHERE ai.attempt_id = $1
  AND ai.organization_id = $2
ORDER BY ai.position;

-- name: GetAttemptItemForGrading :one
SELECT
    ai.id,
    ai.organization_id,
    ai.attempt_id,
    ai.question_version_id,
    ai.position,
    ai.points::text AS points,
    ai.question_type,
    aa.answer_payload
FROM attempt_items ai
LEFT JOIN attempt_answers aa
    ON aa.attempt_item_id = ai.id
    AND aa.organization_id = ai.organization_id
    AND aa.attempt_id = ai.attempt_id
WHERE ai.id = $1
  AND ai.organization_id = $2
LIMIT 1;

-- name: UpsertItemGrade :one
INSERT INTO item_grades (
    organization_id,
    attempt_id,
    attempt_item_id,
    grader_user_id,
    awarded_score,
    feedback,
    graded_at,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, now(), now(), now())
ON CONFLICT (organization_id, attempt_item_id) DO UPDATE SET
    grader_user_id = EXCLUDED.grader_user_id,
    awarded_score = EXCLUDED.awarded_score,
    feedback = EXCLUDED.feedback,
    graded_at = now(),
    updated_at = now()
RETURNING id, organization_id, attempt_id, attempt_item_id, grader_user_id, awarded_score::text AS awarded_score, feedback, graded_at, created_at, updated_at;

-- name: GetItemGrade :one
SELECT
    id, organization_id, attempt_id, attempt_item_id, grader_user_id,
    awarded_score::text AS awarded_score, COALESCE(feedback, '') AS feedback, graded_at, created_at, updated_at
FROM item_grades
WHERE organization_id = $1
  AND attempt_item_id = $2
LIMIT 1;

-- name: GetItemGradeByID :one
SELECT
    id, organization_id, attempt_id, attempt_item_id, grader_user_id,
    awarded_score::text AS awarded_score, COALESCE(feedback, '') AS feedback, graded_at, created_at, updated_at
FROM item_grades
WHERE id = $1
  AND organization_id = $2
LIMIT 1;

-- name: RecomputeAttemptScore :one
-- Recompute the attempt's score from the sum of item-level grades and
-- promote the attempt to GRADED only when every non-MCQ item has a grade.
WITH per_item AS (
    SELECT
        ai.id,
        ai.question_type,
        ai.points::numeric AS points,
        ig.id AS grade_id
    FROM attempt_items ai
    LEFT JOIN item_grades ig
        ON ig.attempt_item_id = ai.id
        AND ig.organization_id = ai.organization_id
    WHERE ai.attempt_id = $1
      AND ai.organization_id = $2
),
auto_grade AS (
    -- MCQ auto-grade on submit is preserved; we don't recompute MCQ correctness
    -- here. If a manual item_grade exists for an MCQ we trust the awarded_score.
    SELECT
        COALESCE(SUM(COALESCE(ig.awarded_score, 0)), 0)::numeric AS score,
        COALESCE(SUM(pi.points), 0)::numeric AS max_score,
        BOOL_OR(
            pi.question_type IN ('essay', 'short_answer') AND pi.grade_id IS NULL
        ) AS has_pending
    FROM per_item pi
    LEFT JOIN item_grades ig
        ON ig.attempt_item_id = pi.id
        AND ig.organization_id = $2
)
UPDATE attempts
SET score = ag.score,
    max_score = ag.max_score,
    grading_status = CASE WHEN ag.has_pending THEN 'PENDING_REVIEW' ELSE 'GRADED' END,
    updated_at = now()
FROM auto_grade ag
WHERE attempts.id = $1
  AND attempts.organization_id = $2
RETURNING attempts.id, attempts.score::text AS score, attempts.max_score::text AS max_score, attempts.grading_status;
