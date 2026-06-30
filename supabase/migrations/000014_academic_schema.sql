CREATE TABLE IF NOT EXISTS academic_terms (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_academic_terms_org_status
    ON academic_terms (organization_id, status, start_date DESC);

CREATE TABLE IF NOT EXISTS subjects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code text NOT NULL,
    name text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_subjects_org_status
    ON subjects (organization_id, status);

CREATE TABLE IF NOT EXISTS courses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subject_id uuid NOT NULL REFERENCES subjects(id),
    academic_term_id uuid NOT NULL REFERENCES academic_terms(id),
    code text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_courses_org_term_status
    ON courses (organization_id, academic_term_id, status);

CREATE TABLE IF NOT EXISTS class_sections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    course_id uuid NOT NULL REFERENCES courses(id),
    name text NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_class_sections_org_course_status
    ON class_sections (organization_id, course_id, status);

CREATE TABLE IF NOT EXISTS class_teachers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    class_section_id uuid NOT NULL REFERENCES class_sections(id) ON DELETE CASCADE,
    membership_id uuid NOT NULL REFERENCES organization_memberships(id),
    role text NOT NULL DEFAULT 'teacher' CHECK (role IN ('teacher', 'assistant')),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (class_section_id, membership_id)
);

CREATE INDEX IF NOT EXISTS idx_class_teachers_class
    ON class_teachers (class_section_id, status);
CREATE INDEX IF NOT EXISTS idx_class_teachers_membership
    ON class_teachers (membership_id, status);

CREATE TABLE IF NOT EXISTS enrollments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    class_section_id uuid NOT NULL REFERENCES class_sections(id) ON DELETE CASCADE,
    membership_id uuid NOT NULL REFERENCES organization_memberships(id),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (class_section_id, membership_id)
);

CREATE INDEX IF NOT EXISTS idx_enrollments_class
    ON enrollments (class_section_id, status);
CREATE INDEX IF NOT EXISTS idx_enrollments_membership
    ON enrollments (membership_id, status);

ALTER TABLE assessments
    ADD COLUMN IF NOT EXISTS class_section_id uuid REFERENCES class_sections(id);

CREATE INDEX IF NOT EXISTS idx_assessments_class_section
    ON assessments (class_section_id) WHERE class_section_id IS NOT NULL;
