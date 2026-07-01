-- name: CreateResource :one
INSERT INTO resources (
  organization_id, title, description, context_type, context_id, status, created_by
) VALUES (
  $1, $2, $3, $4::resource_context_type, $5, 'DRAFT', $6
)
RETURNING id, organization_id, title, description, context_type::text, context_id, status::text, created_by, created_at, updated_at, published_at;

-- name: ListResources :many
SELECT
  id, organization_id, title, description, context_type::text, context_id, status::text,
  created_by, created_at, updated_at, published_at
FROM resources
WHERE organization_id = $1
  AND status::text = ANY($2::text[])
ORDER BY updated_at DESC;

-- name: GetResource :one
SELECT
  id, organization_id, title, description, context_type::text, context_id, status::text,
  created_by, created_at, updated_at, published_at
FROM resources
WHERE id = $1 AND organization_id = $2;

-- name: UpdateResourceStatus :one
UPDATE resources
SET
  status = $3::resource_status,
  published_at = CASE WHEN $3::resource_status = 'PUBLISHED' THEN now() ELSE published_at END,
  updated_at = now()
WHERE id = $1 AND organization_id = $2
RETURNING id, organization_id, title, description, context_type::text, context_id, status::text, created_by, created_at, updated_at, published_at;

-- name: ArchiveResource :exec
UPDATE resources
SET status = 'ARCHIVED', updated_at = now()
WHERE id = $1 AND organization_id = $2;

-- name: CreateResourceFile :one
INSERT INTO resource_files (
  resource_id, organization_id, original_name, storage_key, content_type, size_bytes, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, resource_id, organization_id, original_name, storage_key, content_type, size_bytes, status::text, created_by, created_at;

-- name: ListResourceFiles :many
SELECT
  id, resource_id, organization_id, original_name, storage_key, content_type, size_bytes,
  status::text, created_by, created_at
FROM resource_files
WHERE resource_id = $1 AND organization_id = $2
ORDER BY created_at DESC;

-- name: GetActiveResourceFile :one
SELECT
  id, resource_id, organization_id, original_name, storage_key, content_type, size_bytes,
  status::text, created_by, created_at
FROM resource_files
WHERE resource_id = $1 AND organization_id = $2 AND status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT 1;

-- name: ArchiveResourceFile :exec
UPDATE resource_files
SET status = 'ARCHIVED'
WHERE id = $1 AND organization_id = $2;
