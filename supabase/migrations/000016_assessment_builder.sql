-- 000016_assessment_builder.sql
-- MVP assessment builder schema: sections, builder items, targets, scheduling/revision fields.

ALTER TABLE assessments
ADD COLUMN IF NOT EXISTS revision int NOT NULL DEFAULT 1,
ADD COLUMN IF NOT EXISTS instructions text,
ADD COLUMN IF NOT EXISTS opens_at timestamptz,
ADD COLUMN IF NOT EXISTS closes_at timestamptz;

CREATE TABLE IF NOT EXISTS assessment_sections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    assessment_id uuid NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    title text NOT NULL DEFAULT '',
    position int NOT NULL DEFAULT 0,
    settings_json jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, assessment_id, position)
);

CREATE TABLE IF NOT EXISTS assessment_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    assessment_id uuid NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    assessment_section_id uuid NOT NULL REFERENCES assessment_sections(id) ON DELETE CASCADE,
    question_version_id uuid NOT NULL REFERENCES question_versions(id),
    position int NOT NULL DEFAULT 0,
    points numeric(10,2) NOT NULL DEFAULT '1.00',
    settings_json jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, assessment_section_id, position)
);

CREATE TABLE IF NOT EXISTS assessment_targets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    assessment_id uuid NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    class_section_id uuid NOT NULL REFERENCES class_sections(id),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','ARCHIVED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, assessment_id, class_section_id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_sections_assessment
    ON assessment_sections (organization_id, assessment_id, status);
CREATE INDEX IF NOT EXISTS idx_assessment_items_section
    ON assessment_items (organization_id, assessment_section_id, status);
CREATE INDEX IF NOT EXISTS idx_assessment_items_assessment
    ON assessment_items (organization_id, assessment_id, status);
CREATE INDEX IF NOT EXISTS idx_assessment_targets_assessment
    ON assessment_targets (organization_id, assessment_id, status);

CREATE TRIGGER trg_assessment_sections_updated_at
    BEFORE UPDATE ON assessment_sections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_assessment_items_updated_at
    BEFORE UPDATE ON assessment_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_assessment_targets_updated_at
    BEFORE UPDATE ON assessment_targets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Ensure the demo assessment keeps working with the new columns/defaults.
UPDATE assessments
SET revision = 1
WHERE revision IS NULL;
