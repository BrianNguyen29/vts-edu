WITH org AS (
    SELECT id FROM organizations WHERE code = 'school-a'
),
term AS (
    INSERT INTO academic_terms (organization_id, name, start_date, end_date)
    SELECT org.id, 'Học kỳ 1 2025-2026', '2025-08-01', '2026-01-15'
    FROM org
    RETURNING id
),
subj AS (
    INSERT INTO subjects (organization_id, code, name)
    SELECT org.id, 'MATH', 'Toán'
    FROM org
    RETURNING id
),
course AS (
    INSERT INTO courses (organization_id, subject_id, academic_term_id, code, name)
    SELECT org.id, subj.id, term.id, 'MATH8-HK1', 'Toán 8 - HK1'
    FROM org, term, subj
    RETURNING id
),
cls AS (
    INSERT INTO class_sections (organization_id, course_id, name)
    SELECT org.id, course.id, '8A1'
    FROM org, course
    RETURNING id
),
teacher_m AS (
    SELECT m.id, m.user_id
    FROM organization_memberships m
    JOIN org ON m.organization_id = org.id
    JOIN membership_login_names ln ON ln.organization_id = org.id AND ln.user_id = m.user_id
    WHERE ln.username_normalized = 'gv001'
),
student_m AS (
    SELECT m.id, m.user_id
    FROM organization_memberships m
    JOIN org ON m.organization_id = org.id
    JOIN membership_login_names ln ON ln.organization_id = org.id AND ln.user_id = m.user_id
    WHERE ln.username_normalized = 'hs001'
),
ct AS (
    INSERT INTO class_teachers (organization_id, class_section_id, membership_id, role)
    SELECT org.id, cls.id, teacher_m.id, 'teacher'
    FROM org, cls, teacher_m
    RETURNING id
),
en AS (
    INSERT INTO enrollments (organization_id, class_section_id, membership_id)
    SELECT org.id, cls.id, student_m.id
    FROM org, cls, student_m
    RETURNING id
)
SELECT * FROM ct, en;
