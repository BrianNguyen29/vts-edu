-- name: GetAttempt :one
SELECT
    a.id,
    a.organization_id,
    a.assessment_id,
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
  AND a.student_user_id = $3
LIMIT 1;

-- name: GetAttemptItems :many
SELECT
    ai.id,
    ai.question_version_id,
    ai.position,
    ai.points::text,
    ai.prompt_json,
    ai.choices_json,
    aa.answer_payload,
    ai.answer_key_json,
    ai.question_type,
    aa.revision,
    aa.answered_at
FROM attempt_items ai
LEFT JOIN attempt_answers aa
    ON aa.attempt_item_id = ai.id
    AND aa.organization_id = ai.organization_id
    AND aa.attempt_id = ai.attempt_id
WHERE ai.attempt_id = $1
  AND ai.organization_id = $2
ORDER BY ai.position;

-- name: GetAttemptForUpdate :one
SELECT
    id,
    organization_id,
    assessment_id,
    publication_id,
    status,
    started_at,
    expires_at,
    submitted_at,
    score,
    max_score,
    grading_status
FROM attempts
WHERE id = $1
  AND organization_id = $2
  AND student_user_id = $3
FOR UPDATE;

-- name: ItemExists :one
SELECT EXISTS (
    SELECT 1
    FROM attempt_items
    WHERE id = $1
      AND attempt_id = $2
      AND organization_id = $3
);

-- name: UpsertAnswer :one
INSERT INTO attempt_answers (
    organization_id,
    attempt_id,
    attempt_item_id,
    answer_payload,
    revision,
    answered_at,
    updated_at
) VALUES ($1, $2, $3, $4, 1, now(), now())
ON CONFLICT (organization_id, attempt_id, attempt_item_id)
DO UPDATE SET
    answer_payload = EXCLUDED.answer_payload,
    revision = attempt_answers.revision + 1,
    answered_at = now(),
    updated_at = now()
RETURNING revision, answered_at, answer_payload;

-- name: MarkAttemptExpired :exec
UPDATE attempts
SET status = 'EXPIRED', updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND student_user_id = $3;

-- name: SubmitAttempt :one
UPDATE attempts
SET status = 'SUBMITTED',
    submitted_at = now(),
    score = $4,
    max_score = $5,
    grading_status = $6,
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND student_user_id = $3
  AND status = 'IN_PROGRESS'
RETURNING submitted_at, COALESCE(score, '0')::text AS score, COALESCE(max_score, '0')::text AS max_score, grading_status;

-- name: ListAssignedAssessments :many
SELECT a.id, a.title, a.status, a.duration_minutes, a.max_attempts, a.revision, a.opens_at, a.closes_at, ap.id AS publication_id, ap.published_at, COALESCE(used.cnt, 0) AS attempts_used
FROM assessments a
JOIN assessment_targets t ON t.assessment_id = a.id AND t.status = 'ACTIVE'
JOIN class_sections cs ON cs.id = t.class_section_id AND cs.status = 'ACTIVE'
JOIN enrollments e ON e.class_section_id = t.class_section_id AND e.status = 'ACTIVE'
JOIN organization_memberships m ON m.id = e.membership_id AND m.user_id = $1 AND m.organization_id = $2 AND m.status = 'ACTIVE'
LEFT JOIN LATERAL (
    SELECT id, published_at
    FROM assessment_publications
    WHERE assessment_id = a.id AND organization_id = $2
    ORDER BY version DESC
    LIMIT 1
) ap ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS cnt
    FROM attempts att
    WHERE att.organization_id = $2
      AND att.student_user_id = $1
      AND att.assessment_id = a.id
) used ON true
WHERE a.organization_id = $2
  AND a.status IN ('OPEN', 'PUBLISHED')
  AND ap.id IS NOT NULL
ORDER BY a.created_at DESC;

-- name: ListStudentAttempts :many
SELECT a.id, a.assessment_id, a.publication_id, a.status, a.started_at, a.expires_at, a.submitted_at, a.created_at, CASE WHEN a.score IS NULL THEN ''::text ELSE a.score::text END AS score, CASE WHEN a.max_score IS NULL THEN ''::text ELSE a.max_score::text END AS max_score, a.grading_status, asmt.title AS assessment_title
FROM attempts a
JOIN assessments asmt ON asmt.id = a.assessment_id AND asmt.organization_id = a.organization_id
WHERE a.organization_id = $1
  AND a.student_user_id = $2
  AND (sqlc.arg(cursor_key)::text = '' OR a.created_at < sqlc.arg(cursor_key)::timestamptz OR (a.created_at = sqlc.arg(cursor_key)::timestamptz AND a.id::text < sqlc.arg(cursor_id)))
ORDER BY a.created_at DESC, a.id DESC
LIMIT sqlc.arg(page_limit)::int;

-- name: GetLatestPublication :one
SELECT id, snapshot_json, published_at
FROM assessment_publications
WHERE organization_id = $1
  AND assessment_id = $2
ORDER BY version DESC
LIMIT 1;

-- name: GetQuestionVersionType :one
SELECT question_type
FROM question_versions
WHERE id = $1;

-- name: ListQuestionVersionTypes :many
SELECT id, question_type
FROM question_versions
WHERE id = ANY($1::uuid[]);

-- name: GetInProgressAttempt :one
SELECT id, organization_id, assessment_id, publication_id, status, started_at, expires_at, submitted_at, score, max_score, grading_status
FROM attempts
WHERE organization_id = $1
  AND student_user_id = $2
  AND assessment_id = $3
  AND status = 'IN_PROGRESS'
LIMIT 1;

-- name: CountStudentAttempts :one
SELECT COUNT(*)
FROM attempts
WHERE organization_id = $1
  AND student_user_id = $2
  AND assessment_id = $3;

-- name: CreateAttempt :one
INSERT INTO attempts (organization_id, assessment_id, student_user_id, publication_id, status, started_at, expires_at)
VALUES ($1, $2, $3, $4, 'IN_PROGRESS', $5, $6)
RETURNING id, organization_id, assessment_id, publication_id, status, started_at, expires_at, submitted_at, score, max_score, grading_status;

-- name: CreateAttemptItem :exec
INSERT INTO attempt_items (organization_id, attempt_id, question_version_id, position, points, prompt_json, choices_json, answer_key_json, question_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
