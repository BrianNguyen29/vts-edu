-- name: ListPublishedByOrganization :many
SELECT id, title, status, duration_minutes, created_at
FROM assessments
WHERE organization_id = sqlc.arg(organization_id)
  AND status IN ('OPEN', 'PUBLISHED')
  AND (sqlc.arg(search_query)::text = '' OR title ILIKE '%' || sqlc.arg(search_query) || '%')
  AND (sqlc.arg(cursor_key)::text = '' OR created_at < sqlc.arg(cursor_key)::timestamptz OR (created_at = sqlc.arg(cursor_key)::timestamptz AND id::text < sqlc.arg(cursor_id)))
ORDER BY created_at DESC, id DESC
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: CountPublishedByOrganization :one
SELECT COUNT(*)
FROM assessments
WHERE organization_id = sqlc.arg(organization_id)
  AND status IN ('OPEN', 'PUBLISHED')
  AND (sqlc.arg(search_query)::text = '' OR title ILIKE '%' || sqlc.arg(search_query) || '%');

-- Assessment builder queries

-- name: CreateAssessment :one
INSERT INTO assessments (organization_id, class_section_id, title, duration_minutes, max_attempts, status)
VALUES (sqlc.arg(organization_id), sqlc.arg(class_section_id), sqlc.arg(title), sqlc.arg(duration_minutes), sqlc.arg(max_attempts), 'DRAFT')
RETURNING id, organization_id, class_section_id, title, duration_minutes, max_attempts, settings_json, status, revision, instructions, opens_at, closes_at, created_at, updated_at;

-- name: ListAssessmentsByClass :many
SELECT a.id, a.title, a.status, a.duration_minutes, a.revision, a.opens_at, a.closes_at, a.created_at
FROM assessments a
JOIN assessment_targets t ON t.assessment_id = a.id AND t.status = 'ACTIVE'
WHERE a.organization_id = sqlc.arg(organization_id)
  AND t.class_section_id = sqlc.arg(class_section_id)
  AND a.status != 'ARCHIVED'
ORDER BY a.created_at DESC;

-- name: GetAssessment :one
SELECT id, organization_id, class_section_id, title, duration_minutes, max_attempts, settings_json, status, revision, instructions, opens_at, closes_at, created_at, updated_at
FROM assessments
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: UpdateAssessmentSettings :execrows
UPDATE assessments
SET title = sqlc.arg(title),
    duration_minutes = sqlc.arg(duration_minutes),
    max_attempts = sqlc.arg(max_attempts),
    instructions = sqlc.arg(instructions),
    opens_at = sqlc.arg(opens_at),
    closes_at = sqlc.arg(closes_at),
    settings_json = sqlc.arg(settings_json),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'DRAFT';

-- name: PublishAssessment :execrows
UPDATE assessments
SET status = sqlc.arg(new_status),
    revision = revision + 1,
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'DRAFT';

-- name: GetAssessmentRevision :one
SELECT revision
FROM assessments
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: InsertAssessmentPublication :one
INSERT INTO assessment_publications (organization_id, assessment_id, version, snapshot_json, published_at)
VALUES (sqlc.arg(organization_id), sqlc.arg(assessment_id), sqlc.arg(version), sqlc.arg(snapshot_json), now())
RETURNING id, version, snapshot_json, published_at;

-- name: CreateAssessmentSection :one
INSERT INTO assessment_sections (organization_id, assessment_id, title, position)
VALUES (sqlc.arg(organization_id), sqlc.arg(assessment_id), sqlc.arg(title), sqlc.arg(position))
RETURNING id, assessment_id, title, position, settings_json, status;

-- name: CreateAssessmentItem :one
INSERT INTO assessment_items (organization_id, assessment_id, assessment_section_id, question_version_id, position, points)
VALUES (sqlc.arg(organization_id), sqlc.arg(assessment_id), sqlc.arg(assessment_section_id), sqlc.arg(question_version_id), sqlc.arg(position), sqlc.arg(points))
RETURNING id, assessment_section_id, question_version_id, position, points, status;

-- name: CreateAssessmentTarget :one
INSERT INTO assessment_targets (organization_id, assessment_id, class_section_id)
VALUES (sqlc.arg(organization_id), sqlc.arg(assessment_id), sqlc.arg(class_section_id))
RETURNING id, assessment_id, class_section_id, status;

-- name: GetSectionAssessmentID :one
SELECT assessment_id
FROM assessment_sections
WHERE id = sqlc.arg(section_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: GetAssessmentSections :many
SELECT id, assessment_id, title, position, settings_json, status
FROM assessment_sections
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE'
ORDER BY position;

-- name: GetAssessmentItems :many
SELECT id, assessment_section_id, question_version_id, position, points, status
FROM assessment_items
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE'
ORDER BY assessment_section_id, position;

-- name: GetAssessmentItemsBySection :many
SELECT id, assessment_section_id, question_version_id, position, points, status
FROM assessment_items
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_section_id = sqlc.arg(section_id)
  AND status = 'ACTIVE'
ORDER BY position;

-- name: GetAssessmentItemsWithContent :many
SELECT ai.id, ai.assessment_section_id, ai.question_version_id, ai.position, ai.points,
       qv.prompt_json, qv.choices_json, qv.answer_key_json, qv.max_score, qv.question_type
FROM assessment_items ai
JOIN question_versions qv ON qv.id = ai.question_version_id
WHERE ai.organization_id = sqlc.arg(organization_id)
  AND ai.assessment_id = sqlc.arg(assessment_id)
  AND ai.status = 'ACTIVE'
ORDER BY ai.assessment_section_id, ai.position;

-- name: GetAssessmentTargets :many
SELECT id, assessment_id, class_section_id, status
FROM assessment_targets
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE';

-- name: CountAssessmentSections :one
SELECT COUNT(*)
FROM assessment_sections
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE';

-- name: CountAssessmentItems :one
SELECT COUNT(*)
FROM assessment_items
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE';

-- name: CountAssessmentTargets :one
SELECT COUNT(*)
FROM assessment_targets
WHERE organization_id = sqlc.arg(organization_id)
  AND assessment_id = sqlc.arg(assessment_id)
  AND status = 'ACTIVE';

-- name: IsClassManager :one
SELECT EXISTS (
    SELECT 1
    FROM organization_memberships m
    JOIN membership_roles mr ON mr.membership_id = m.id
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.user_id = sqlc.arg(user_id)
      AND m.status = 'ACTIVE'
      AND mr.role = 'admin'
) OR EXISTS (
    SELECT 1
    FROM organization_memberships m
    JOIN class_teachers ct ON ct.membership_id = m.id AND ct.status = 'ACTIVE'
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.user_id = sqlc.arg(user_id)
      AND m.status = 'ACTIVE'
      AND ct.class_section_id = sqlc.arg(class_section_id)
);

-- name: IsAssessmentManager :one
SELECT EXISTS (
    SELECT 1
    FROM organization_memberships m
    JOIN membership_roles mr ON mr.membership_id = m.id
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.user_id = sqlc.arg(user_id)
      AND m.status = 'ACTIVE'
      AND mr.role = 'admin'
) OR EXISTS (
    SELECT 1
    FROM organization_memberships m
    JOIN class_teachers ct ON ct.membership_id = m.id AND ct.status = 'ACTIVE'
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.user_id = sqlc.arg(user_id)
      AND m.status = 'ACTIVE'
      AND (
          ct.class_section_id = (SELECT a.class_section_id FROM assessments a WHERE a.id = sqlc.arg(assessment_id) AND a.organization_id = sqlc.arg(organization_id))
          OR ct.class_section_id IN (
              SELECT class_section_id FROM assessment_targets
              WHERE assessment_id = sqlc.arg(assessment_id) AND organization_id = sqlc.arg(organization_id) AND status = 'ACTIVE'
          )
      )
);

-- name: QuestionVersionExists :one
SELECT EXISTS (
    SELECT 1
    FROM question_versions qv
    JOIN questions q ON q.id = qv.question_id
    JOIN question_banks qb ON qb.id = q.question_bank_id
    WHERE qv.id = sqlc.arg(question_version_id)
      AND qb.organization_id = sqlc.arg(organization_id)
      AND qv.status = 'PUBLISHED'
);

-- name: GetAssessmentSection :one
SELECT id, assessment_id, title, position, settings_json, status
FROM assessment_sections
WHERE id = sqlc.arg(section_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: UpdateAssessmentSection :one
UPDATE assessment_sections
SET title = COALESCE(NULLIF(sqlc.arg(title)::text, ''), title),
    position = sqlc.arg(position),
    updated_at = now()
WHERE id = sqlc.arg(section_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE'
RETURNING id, assessment_id, title, position, settings_json, status;

-- name: ArchiveAssessmentSection :execrows
UPDATE assessment_sections
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = sqlc.arg(section_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE';

-- name: GetAssessmentItem :one
SELECT id, assessment_section_id, question_version_id, position, points, status
FROM assessment_items
WHERE id = sqlc.arg(item_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: GetItemAssessmentID :one
SELECT assessment_id
FROM assessment_items
WHERE id = sqlc.arg(item_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: UpdateAssessmentItem :one
UPDATE assessment_items
SET question_version_id = COALESCE(NULLIF(sqlc.arg(question_version_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid), question_version_id),
    points = COALESCE(NULLIF(sqlc.arg(points)::numeric, '0'::numeric), points),
    position = sqlc.arg(position),
    updated_at = now()
WHERE id = sqlc.arg(item_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE'
RETURNING id, assessment_section_id, question_version_id, position, points, status;

-- name: ArchiveAssessmentItem :execrows
UPDATE assessment_items
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = sqlc.arg(item_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE';

-- name: GetAssessmentTarget :one
SELECT id, assessment_id, class_section_id, status
FROM assessment_targets
WHERE id = sqlc.arg(target_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: GetTargetAssessmentID :one
SELECT assessment_id
FROM assessment_targets
WHERE id = sqlc.arg(target_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: ArchiveAssessmentTarget :execrows
UPDATE assessment_targets
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = sqlc.arg(target_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE';

-- name: UpdateAssessmentSectionPosition :execrows
UPDATE assessment_sections
SET position = sqlc.arg(position),
    updated_at = now()
WHERE id = sqlc.arg(section_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE';

-- name: UpdateAssessmentItemPosition :execrows
UPDATE assessment_items
SET position = sqlc.arg(position),
    updated_at = now()
WHERE id = sqlc.arg(item_id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'ACTIVE';

-- name: ListQuestions :many
SELECT q.id, q.question_bank_id, COALESCE(qv.id, '00000000-0000-0000-0000-000000000000'::uuid) AS question_version_id, COALESCE(qv.status, '') AS question_version_status, COALESCE(qv.question_type, '') AS question_type, qv.prompt_json ->> 'text' AS prompt_text
FROM questions q
JOIN question_banks qb ON qb.id = q.question_bank_id
LEFT JOIN LATERAL (
    SELECT id, status, prompt_json, question_type
    FROM question_versions
    WHERE question_id = q.id
      AND status = 'PUBLISHED'
    ORDER BY version DESC, created_at DESC
    LIMIT 1
) qv ON true
WHERE qb.organization_id = sqlc.arg(organization_id)
  AND q.status = 'ACTIVE'
  AND (sqlc.arg(bank_id)::text = '' OR q.question_bank_id::text = sqlc.arg(bank_id)::text)
  AND (sqlc.arg(search_query)::text = '' OR qv.prompt_json ->> 'text' ILIKE '%' || sqlc.arg(search_query) || '%')
ORDER BY q.created_at DESC
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: CountQuestions :one
SELECT COUNT(*)
FROM questions q
JOIN question_banks qb ON qb.id = q.question_bank_id
LEFT JOIN LATERAL (
    SELECT id, status, prompt_json
    FROM question_versions
    WHERE question_id = q.id
      AND status = 'PUBLISHED'
    ORDER BY version DESC, created_at DESC
    LIMIT 1
) qv ON true
WHERE qb.organization_id = sqlc.arg(organization_id)
  AND q.status = 'ACTIVE'
  AND (sqlc.arg(bank_id)::text = '' OR q.question_bank_id::text = sqlc.arg(bank_id)::text)
  AND (sqlc.arg(search_query)::text = '' OR qv.prompt_json ->> 'text' ILIKE '%' || sqlc.arg(search_query) || '%');

-- name: ListAssessmentPublications :many
SELECT ap.id, ap.version, a.status, ap.published_at
FROM assessment_publications ap
JOIN assessments a ON a.id = ap.assessment_id AND a.organization_id = ap.organization_id
WHERE ap.organization_id = sqlc.arg(organization_id)
  AND ap.assessment_id = sqlc.arg(assessment_id)
ORDER BY ap.version DESC;

-- name: IsQuestionVersionPublished :one
SELECT EXISTS (
    SELECT 1
    FROM question_versions qv
    JOIN questions q ON q.id = qv.question_id
    JOIN question_banks qb ON qb.id = q.question_bank_id
    WHERE qv.id = sqlc.arg(question_version_id)
      AND qb.organization_id = sqlc.arg(organization_id)
      AND qv.status = 'PUBLISHED'
);

-- name: IsClassSectionActive :one
SELECT EXISTS (
    SELECT 1
    FROM class_sections
    WHERE id = sqlc.arg(class_section_id)
      AND organization_id = sqlc.arg(organization_id)
      AND status = 'ACTIVE'
);

-- Scheduler: transition published/scheduled assessments to open when opens_at passes.
-- name: TransitionAssessmentsToOpen :execrows
UPDATE assessments
SET status = 'OPEN',
    updated_at = now()
WHERE status IN ('SCHEDULED', 'PUBLISHED')
  AND opens_at IS NOT NULL
  AND opens_at <= now();

-- Scheduler: transition open assessments to closed when closes_at passes.
-- name: TransitionAssessmentsToClosed :execrows
UPDATE assessments
SET status = 'CLOSED',
    updated_at = now()
WHERE status = 'OPEN'
  AND closes_at IS NOT NULL
  AND closes_at <= now();

-- Question bank editor queries

-- name: CreateQuestionBank :one
INSERT INTO question_banks (organization_id, title)
VALUES (sqlc.arg(organization_id), sqlc.arg(title))
RETURNING id, organization_id, title, status, created_at, updated_at;

-- name: ListQuestionBanksByOrganization :many
SELECT id, organization_id, title, status, created_at, updated_at
FROM question_banks
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.arg(include_archived)::bool OR status = 'ACTIVE')
ORDER BY created_at DESC
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: GetQuestionBank :one
SELECT id, organization_id, title, status, created_at, updated_at
FROM question_banks
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: CreateQuestion :one
INSERT INTO questions (question_bank_id)
VALUES (sqlc.arg(question_bank_id))
RETURNING id, question_bank_id, status, created_at, updated_at;

-- name: ListQuestionsInBank :many
SELECT q.id, q.question_bank_id, q.status, q.created_at, q.updated_at,
       qv.id AS latest_version_id, qv.status AS latest_version_status,
       qv.question_type, qv.version AS latest_version
FROM questions q
LEFT JOIN LATERAL (
    SELECT id, status, question_type, version
    FROM question_versions
    WHERE question_id = q.id
    ORDER BY version DESC, created_at DESC
    LIMIT 1
) qv ON true
WHERE q.question_bank_id = sqlc.arg(bank_id)
  AND (sqlc.arg(include_archived)::bool OR q.status = 'ACTIVE')
ORDER BY q.created_at DESC
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: GetQuestion :one
SELECT id, question_bank_id, status, created_at, updated_at
FROM questions
WHERE id = sqlc.arg(id)
  AND question_bank_id = sqlc.arg(bank_id);

-- name: GetQuestionWithBank :one
SELECT q.id, q.question_bank_id, qb.organization_id, q.status, q.created_at, q.updated_at
FROM questions q
JOIN question_banks qb ON qb.id = q.question_bank_id
WHERE q.id = sqlc.arg(id);

-- name: CreateQuestionVersion :one
INSERT INTO question_versions (
    question_id, version, prompt_json, choices_json, answer_key_json, max_score, status, question_type
) VALUES (
    sqlc.arg(question_id),
    sqlc.arg(version),
    sqlc.arg(prompt_json),
    sqlc.arg(choices_json),
    sqlc.arg(answer_key_json),
    sqlc.arg(max_score),
    sqlc.arg(status),
    sqlc.arg(question_type)
)
RETURNING id, question_id, version, prompt_json, choices_json, answer_key_json, max_score, status, question_type, created_at;

-- name: GetQuestionVersion :one
SELECT qv.id, qv.question_id, qv.version, qv.prompt_json, qv.choices_json, qv.answer_key_json, qv.max_score, qv.status, qv.question_type, qv.created_at
FROM question_versions qv
JOIN questions q ON q.id = qv.question_id
JOIN question_banks qb ON qb.id = q.question_bank_id
WHERE qv.id = sqlc.arg(version_id)
  AND qb.organization_id = sqlc.arg(organization_id);

-- name: GetLatestVersionNumber :one
SELECT COALESCE(MAX(version), 0)::int AS version
FROM question_versions
WHERE question_id = sqlc.arg(question_id);

-- name: PublishQuestionVersion :one
UPDATE question_versions
SET status = 'PUBLISHED',
    created_at = created_at
WHERE id = sqlc.arg(version_id)
  AND status = 'DRAFT'
RETURNING id, question_id, version, prompt_json, choices_json, answer_key_json, max_score, status, question_type, created_at;
