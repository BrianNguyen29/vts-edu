-- name: ListUsers :many
SELECT
    u.id,
    u.display_name,
    u.email,
    ln.username_normalized,
    u.must_change_password,
    array_agg(mr.role) FILTER (WHERE mr.role IS NOT NULL)
FROM users u
JOIN organization_memberships m ON m.user_id = u.id
JOIN membership_login_names ln ON ln.user_id = u.id AND ln.organization_id = m.organization_id
LEFT JOIN membership_roles mr ON mr.membership_id = m.id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND m.status = 'ACTIVE'
  AND ln.status = 'ACTIVE'
  AND (sqlc.arg(search_query)::text = '' OR ln.username_normalized ILIKE '%' || sqlc.arg(search_query) || '%' OR u.display_name ILIKE '%' || sqlc.arg(search_query) || '%' OR u.email ILIKE '%' || sqlc.arg(search_query) || '%')
  AND (sqlc.arg(cursor_key)::text = '' OR ln.username_normalized > sqlc.arg(cursor_key) OR (ln.username_normalized = sqlc.arg(cursor_key) AND u.id::text > sqlc.arg(cursor_id)))
GROUP BY u.id, u.display_name, u.email, ln.username_normalized, u.must_change_password
ORDER BY ln.username_normalized, u.id
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users u
JOIN organization_memberships m ON m.user_id = u.id
JOIN membership_login_names ln ON ln.user_id = u.id AND ln.organization_id = m.organization_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND m.status = 'ACTIVE'
  AND ln.status = 'ACTIVE'
  AND (sqlc.arg(search_query)::text = '' OR ln.username_normalized ILIKE '%' || sqlc.arg(search_query) || '%' OR u.display_name ILIKE '%' || sqlc.arg(search_query) || '%' OR u.email ILIKE '%' || sqlc.arg(search_query) || '%');

-- name: ListAuditLogs :many
SELECT
    id,
    actor_user_id,
    action,
    resource_type,
    resource_id,
    before_json,
    after_json,
    metadata_json,
    created_at
FROM audit_logs
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.arg(action_name)::text = '' OR action = sqlc.arg(action_name))
  AND (sqlc.arg(actor_user_id)::text = '' OR actor_user_id::text = sqlc.arg(actor_user_id))
  AND (sqlc.arg(from_time)::text = '' OR created_at >= sqlc.arg(from_time)::timestamptz)
  AND (sqlc.arg(to_time)::text = '' OR created_at <= sqlc.arg(to_time)::timestamptz)
  AND (sqlc.arg(cursor_key)::text = '' OR created_at < sqlc.arg(cursor_key)::timestamptz OR (created_at = sqlc.arg(cursor_key)::timestamptz AND id::text < sqlc.arg(cursor_id)))
ORDER BY created_at DESC, id DESC
LIMIT NULLIF(sqlc.arg(page_limit)::int, 0) OFFSET sqlc.arg(page_offset)::int;

-- name: CountAuditLogs :one
SELECT COUNT(*)
FROM audit_logs
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.arg(action_name)::text = '' OR action = sqlc.arg(action_name))
  AND (sqlc.arg(actor_user_id)::text = '' OR actor_user_id::text = sqlc.arg(actor_user_id))
  AND (sqlc.arg(from_time)::text = '' OR created_at >= sqlc.arg(from_time)::timestamptz)
  AND (sqlc.arg(to_time)::text = '' OR created_at <= sqlc.arg(to_time)::timestamptz);

-- name: ExportAuditLogs :many
SELECT
    al.id,
    al.created_at,
    u.display_name AS actor_name,
    al.actor_user_id,
    al.action,
    al.resource_type,
    al.resource_id,
    al.before_json,
    al.after_json,
    al.metadata_json
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_user_id
WHERE al.organization_id = sqlc.arg(organization_id)
  AND (sqlc.arg(action_name)::text = '' OR al.action = sqlc.arg(action_name))
  AND (sqlc.arg(actor_user_id)::text = '' OR al.actor_user_id::text = sqlc.arg(actor_user_id))
  AND (sqlc.arg(from_time)::text = '' OR al.created_at >= sqlc.arg(from_time)::timestamptz)
  AND (sqlc.arg(to_time)::text = '' OR al.created_at <= sqlc.arg(to_time)::timestamptz)
ORDER BY al.created_at DESC, al.id DESC;

-- name: LoginExists :one
SELECT EXISTS (
    SELECT 1
    FROM membership_login_names
    WHERE organization_id = $1
      AND lower(username_normalized) = lower($2)
);

-- name: CreateUser :one
INSERT INTO users (display_name, email, must_change_password)
VALUES ($1, $2, true)
RETURNING id;

-- name: CreateMembership :one
INSERT INTO organization_memberships (organization_id, user_id)
VALUES ($1, $2)
RETURNING id;

-- name: CreateLoginName :exec
INSERT INTO membership_login_names (organization_id, username_normalized, user_id, password_hash)
VALUES ($1, lower($2), $3, $4);

-- name: CreateRole :exec
INSERT INTO membership_roles (membership_id, role)
VALUES ($1, $2);

-- name: GetMembershipID :one
SELECT id
FROM organization_memberships
WHERE organization_id = $1
  AND user_id = $2
  AND status = 'ACTIVE'
LIMIT 1;

-- name: DeleteRoles :exec
DELETE FROM membership_roles
WHERE membership_id = $1;

-- name: BumpAuthVersion :exec
UPDATE users
SET auth_version = auth_version + 1
WHERE id = $1;

-- name: GetLoginPasswordHash :one
SELECT password_hash
FROM membership_login_names
WHERE user_id = $1
  AND organization_id = $2
LIMIT 1;

-- name: ResetPassword :execrows
UPDATE membership_login_names
SET password_hash = $1
WHERE user_id = $2
  AND organization_id = $3;

-- name: SetMustChangePassword :exec
UPDATE users
SET must_change_password = true,
    auth_version = auth_version + 1
WHERE id = $1;

-- name: RevokeUserSessions :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;

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

-- name: GetOrganization :one
SELECT id, code, name
FROM organizations
WHERE id = $1
LIMIT 1;

-- name: UpdateOrganization :execrows
UPDATE organizations
SET name = $1
WHERE id = $2;

-- name: ListPasswordHistory :many
SELECT password_hash
FROM password_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: InsertPasswordHistory :exec
INSERT INTO password_history (user_id, password_hash)
VALUES ($1, $2);

-- name: DeleteOldPasswordHistory :exec
DELETE FROM password_history
WHERE id IN (
    SELECT ph.id
    FROM password_history ph
    WHERE ph.user_id = $1
    ORDER BY ph.created_at DESC
    OFFSET $2
);
