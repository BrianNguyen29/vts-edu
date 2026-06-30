-- name: ListPublishedByOrganization :many
SELECT id, title, status, duration_minutes
FROM assessments
WHERE organization_id = $1
  AND status IN ('OPEN', 'PUBLISHED')
  AND ($2::text = '' OR title ILIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT NULLIF($3::int, 0) OFFSET $4::int;
