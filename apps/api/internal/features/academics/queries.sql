-- name: ListTerms :many
SELECT id, name, start_date, end_date, status
FROM academic_terms
WHERE organization_id = $1
  AND status = 'ACTIVE'
ORDER BY start_date DESC;

-- name: CreateTerm :one
INSERT INTO academic_terms (organization_id, name, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING id, name, start_date, end_date, status;

-- name: UpdateTerm :one
UPDATE academic_terms
SET name = $3,
    start_date = $4,
    end_date = $5,
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE'
RETURNING id, name, start_date, end_date, status;

-- name: ArchiveTerm :execrows
UPDATE academic_terms
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE';

-- name: ListSubjects :many
SELECT id, code, name, description, status
FROM subjects
WHERE organization_id = $1
  AND status = 'ACTIVE'
ORDER BY code;

-- name: CreateSubject :one
INSERT INTO subjects (organization_id, code, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, code, name, description, status;

-- name: UpdateSubject :one
UPDATE subjects
SET code = $3,
    name = $4,
    description = $5,
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE'
RETURNING id, code, name, description, status;

-- name: ArchiveSubject :execrows
UPDATE subjects
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE';

-- name: ListCourses :many
SELECT id, subject_id, academic_term_id, code, name, status
FROM courses
WHERE organization_id = $1
  AND status = 'ACTIVE'
ORDER BY code;

-- name: CreateCourse :one
INSERT INTO courses (organization_id, subject_id, academic_term_id, code, name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, subject_id, academic_term_id, code, name, status;

-- name: UpdateCourse :one
UPDATE courses
SET subject_id = $3,
    academic_term_id = $4,
    code = $5,
    name = $6,
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE'
RETURNING id, subject_id, academic_term_id, code, name, status;

-- name: ArchiveCourse :execrows
UPDATE courses
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE';

-- name: ListClassesAdmin :many
SELECT
    cs.id,
    cs.course_id,
    cs.name,
    (SELECT COUNT(*) FROM enrollments e WHERE e.class_section_id = cs.id AND e.status = 'ACTIVE') AS student_count,
    (SELECT COUNT(*) FROM class_teachers ct WHERE ct.class_section_id = cs.id AND ct.status = 'ACTIVE') AS teacher_count,
    cs.status
FROM class_sections cs
WHERE cs.organization_id = $1
  AND cs.status = 'ACTIVE'
ORDER BY cs.name;

-- name: ListClassesForTeacher :many
SELECT
    cs.id,
    cs.course_id,
    cs.name,
    (SELECT COUNT(*) FROM enrollments e WHERE e.class_section_id = cs.id AND e.status = 'ACTIVE') AS student_count,
    (SELECT COUNT(*) FROM class_teachers ct WHERE ct.class_section_id = cs.id AND ct.status = 'ACTIVE') AS teacher_count,
    cs.status
FROM class_sections cs
JOIN class_teachers ct ON ct.class_section_id = cs.id AND ct.status = 'ACTIVE'
WHERE cs.organization_id = $1
  AND cs.status = 'ACTIVE'
  AND ct.membership_id = $2
ORDER BY cs.name;

-- name: CreateClass :one
INSERT INTO class_sections (organization_id, course_id, name)
VALUES ($1, $2, $3)
RETURNING id, course_id, name, 0::bigint AS student_count, 0::bigint AS teacher_count, status;

-- name: UpdateClass :one
UPDATE class_sections cs
SET course_id = $3,
    name = $4,
    updated_at = now()
WHERE cs.id = $1
  AND cs.organization_id = $2
  AND cs.status = 'ACTIVE'
RETURNING cs.id AS id,
          cs.course_id AS course_id,
          cs.name AS name,
          (SELECT COUNT(*) FROM enrollments e WHERE e.class_section_id = cs.id AND e.status = 'ACTIVE') AS student_count,
          (SELECT COUNT(*) FROM class_teachers ct WHERE ct.class_section_id = cs.id AND ct.status = 'ACTIVE') AS teacher_count,
          cs.status AS status;

-- name: ArchiveClass :execrows
UPDATE class_sections
SET status = 'ARCHIVED',
    updated_at = now()
WHERE id = $1
  AND organization_id = $2
  AND status = 'ACTIVE';

-- name: ListClassTeachers :many
SELECT ct.id, u.id AS user_id, u.display_name, ct.role, ct.status
FROM class_teachers ct
JOIN organization_memberships m ON m.id = ct.membership_id
JOIN users u ON u.id = m.user_id
WHERE ct.organization_id = $1
  AND ct.class_section_id = $2
  AND ct.status = 'ACTIVE'
ORDER BY u.display_name;

-- name: AddClassTeacher :one
WITH inserted AS (
    INSERT INTO class_teachers (organization_id, class_section_id, membership_id, role)
    VALUES ($1, $2, $3, $4)
    RETURNING id, membership_id, role, status
)
SELECT inserted.id, u.id AS user_id, u.display_name, inserted.role, inserted.status
FROM inserted
JOIN organization_memberships m ON m.id = inserted.membership_id
JOIN users u ON u.id = m.user_id;

-- name: RemoveClassTeacher :execrows
UPDATE class_teachers
SET status = 'ARCHIVED',
    updated_at = now()
WHERE organization_id = $1
  AND class_section_id = $2
  AND membership_id = $3
  AND status = 'ACTIVE';

-- name: ListEnrollments :many
SELECT e.id, u.id AS user_id, u.display_name, e.status
FROM enrollments e
JOIN organization_memberships m ON m.id = e.membership_id
JOIN users u ON u.id = m.user_id
WHERE e.organization_id = $1
  AND e.class_section_id = $2
  AND e.status = 'ACTIVE'
ORDER BY u.display_name;

-- name: EnrollStudent :one
WITH inserted AS (
    INSERT INTO enrollments (organization_id, class_section_id, membership_id)
    VALUES ($1, $2, $3)
    RETURNING id, membership_id, status
)
SELECT inserted.id, u.id AS user_id, u.display_name, inserted.status
FROM inserted
JOIN organization_memberships m ON m.id = inserted.membership_id
JOIN users u ON u.id = m.user_id;

-- name: UnenrollStudent :execrows
UPDATE enrollments
SET status = 'ARCHIVED',
    updated_at = now()
WHERE organization_id = $1
  AND class_section_id = $2
  AND membership_id = $3
  AND status = 'ACTIVE';

-- name: GetMembershipByUserID :one
SELECT m.id, m.user_id, array_agg(mr.role) FILTER (WHERE mr.role IS NOT NULL)
FROM organization_memberships m
LEFT JOIN membership_roles mr ON mr.membership_id = m.id
WHERE m.organization_id = $1
  AND m.user_id = $2
  AND m.status = 'ACTIVE'
GROUP BY m.id, m.user_id;

-- name: IsClassTeacher :one
SELECT EXISTS (
    SELECT 1
    FROM class_teachers
    WHERE organization_id = $1
      AND class_section_id = $2
      AND membership_id = $3
      AND status = 'ACTIVE'
);

-- name: ClassExists :one
SELECT EXISTS (
    SELECT 1
    FROM class_sections
    WHERE organization_id = $1
      AND id = $2
      AND status = 'ACTIVE'
);

-- name: IsStudentEnrolled :one
SELECT EXISTS (
    SELECT 1
    FROM enrollments e
    JOIN organization_memberships m ON m.id = e.membership_id
    WHERE e.organization_id = $1
      AND e.class_section_id = $2
      AND m.user_id = $3
      AND e.status = 'ACTIVE'
      AND m.status = 'ACTIVE'
);

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (
    organization_id,
    actor_user_id,
    action,
    resource_type,
    resource_id,
    before_json,
    after_json,
    metadata_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
