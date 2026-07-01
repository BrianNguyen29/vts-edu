CREATE TYPE resource_context_type AS ENUM ('organization', 'class');
CREATE TYPE resource_status AS ENUM ('DRAFT', 'PUBLISHED', 'ARCHIVED');
CREATE TYPE resource_file_status AS ENUM ('ACTIVE', 'ARCHIVED');

CREATE TABLE resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT,
  context_type resource_context_type NOT NULL DEFAULT 'organization',
  context_id UUID NOT NULL,
  status resource_status NOT NULL DEFAULT 'DRAFT',
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ
);

CREATE INDEX idx_resources_org_status ON resources(organization_id, status);
CREATE INDEX idx_resources_context ON resources(context_type, context_id);

CREATE TABLE resource_files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  original_name TEXT NOT NULL,
  storage_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  status resource_file_status NOT NULL DEFAULT 'ACTIVE',
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_resource_files_resource ON resource_files(resource_id);
CREATE INDEX idx_resource_files_org ON resource_files(organization_id);
