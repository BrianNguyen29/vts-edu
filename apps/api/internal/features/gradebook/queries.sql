-- name: AssessmentExists :one
SELECT EXISTS (
    SELECT 1
    FROM assessments
    WHERE id = $1 AND organization_id = $2
);

-- name: ListAssessmentAttempts :many
SELECT
    a.id,
    a.assessment_id,
    a.student_user_id,
    u.display_name AS student_name,
    a.status,
    a.started_at,
    a.expires_at,
    a.submitted_at,
    CASE WHEN a.score IS NULL THEN ''::text ELSE a.score::text END AS score,
    CASE WHEN a.max_score IS NULL THEN ''::text ELSE a.max_score::text END AS max_score,
    a.grading_status
FROM attempts a
JOIN users u ON u.id = a.student_user_id
WHERE a.organization_id = $1
  AND a.assessment_id = $2
ORDER BY a.created_at DESC, a.started_at DESC;

-- name: GetAssessmentResults :one
SELECT
    COUNT(*) AS total_attempts,
    COUNT(*) FILTER (WHERE status = 'SUBMITTED') AS submitted_count,
    COUNT(*) FILTER (WHERE status = 'IN_PROGRESS') AS in_progress_count,
    COUNT(*) FILTER (WHERE status = 'EXPIRED') AS expired_count,
    CASE WHEN COUNT(*) = 0 OR AVG(score) IS NULL THEN ''::text ELSE AVG(score)::text END AS average_score,
    CASE WHEN COUNT(*) = 0 OR MAX(max_score) IS NULL THEN ''::text ELSE MAX(max_score)::text END AS max_score
FROM attempts
WHERE organization_id = $1
  AND assessment_id = $2;

-- name: IsAssessmentTaughtByTeacher :one
SELECT EXISTS (
    SELECT 1
    FROM assessment_targets t
    JOIN class_teachers ct
        ON ct.class_section_id = t.class_section_id
        AND ct.organization_id = t.organization_id
        AND ct.status = 'ACTIVE'
    JOIN organization_memberships m
        ON m.id = ct.membership_id
        AND m.organization_id = t.organization_id
        AND m.status = 'ACTIVE'
    WHERE t.organization_id = $1
      AND t.assessment_id = $2
      AND t.status = 'ACTIVE'
      AND m.user_id = $3
);

-- name: GetClassGradebook :many
SELECT
    u.id AS student_user_id,
    u.display_name AS student_name,
    a.id AS assessment_id,
    a.title AS assessment_title,
    att.id AS attempt_id,
    att.status,
    CASE WHEN att.score IS NULL THEN ''::text ELSE att.score::text END AS score,
    CASE WHEN att.max_score IS NULL THEN ''::text ELSE att.max_score::text END AS max_score,
    att.submitted_at
FROM class_sections cs
JOIN enrollments e
    ON e.class_section_id = cs.id
    AND e.status = 'ACTIVE'
JOIN organization_memberships m
    ON m.id = e.membership_id
    AND m.organization_id = cs.organization_id
    AND m.status = 'ACTIVE'
JOIN users u
    ON u.id = m.user_id
JOIN assessment_targets t
    ON t.class_section_id = cs.id
    AND t.status = 'ACTIVE'
JOIN assessments a
    ON a.id = t.assessment_id
    AND a.organization_id = cs.organization_id
    AND a.status IN ('OPEN', 'PUBLISHED')
LEFT JOIN LATERAL (
    SELECT att.id, att.status, att.score, att.max_score, att.submitted_at
    FROM attempts att
    WHERE att.organization_id = cs.organization_id
      AND att.assessment_id = a.id
      AND att.student_user_id = u.id
    ORDER BY att.created_at DESC, att.started_at DESC
    LIMIT 1
) att ON true
WHERE cs.organization_id = $1
  AND cs.id = $2
ORDER BY u.display_name, a.title;
