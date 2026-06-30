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
