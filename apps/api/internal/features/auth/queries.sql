-- name: FindLoginByCredentials :one
SELECT
    u.id,
    m.id,
    o.id,
    ln.username_normalized,
    ln.password_hash,
    u.auth_version,
    u.must_change_password,
    array_agg(mr.role) FILTER (WHERE mr.role IS NOT NULL)
FROM membership_login_names ln
JOIN organizations o ON o.id = ln.organization_id
JOIN organization_memberships m
    ON m.organization_id = ln.organization_id AND m.user_id = ln.user_id
JOIN users u ON u.id = ln.user_id
LEFT JOIN membership_roles mr ON mr.membership_id = m.id
WHERE lower(o.code) = lower($1)
  AND lower(ln.username_normalized) = lower($2)
  AND o.status = 'ACTIVE'
  AND m.status = 'ACTIVE'
  AND ln.status = 'ACTIVE'
GROUP BY u.id, m.id, o.id, ln.username_normalized, ln.password_hash, u.auth_version, u.must_change_password
LIMIT 1;

-- name: GetLoginByUserID :one
SELECT
    u.id,
    m.id,
    o.id,
    ln.username_normalized,
    ln.password_hash,
    u.auth_version,
    u.must_change_password,
    array_agg(mr.role) FILTER (WHERE mr.role IS NOT NULL)
FROM membership_login_names ln
JOIN organizations o ON o.id = ln.organization_id
JOIN organization_memberships m
    ON m.organization_id = ln.organization_id AND m.user_id = ln.user_id
JOIN users u ON u.id = ln.user_id
LEFT JOIN membership_roles mr ON mr.membership_id = m.id
WHERE u.id = $1
  AND o.id = $2
  AND o.status = 'ACTIVE'
  AND m.status = 'ACTIVE'
  AND ln.status = 'ACTIVE'
GROUP BY u.id, m.id, o.id, ln.username_normalized, ln.password_hash, u.auth_version, u.must_change_password
LIMIT 1;

-- name: GetActorByUserID :one
SELECT ln.user_id, ln.organization_id, ln.username_normalized, u.must_change_password
FROM membership_login_names ln
JOIN users u ON u.id = ln.user_id
WHERE ln.user_id = $1
  AND ln.organization_id = $2
  AND ln.status = 'ACTIVE'
LIMIT 1;

-- name: InsertRefreshSession :one
INSERT INTO refresh_sessions (
    user_id,
    membership_id,
    organization_id,
    token_hash,
    family_id,
    auth_version,
    device_metadata_json,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, '{}', $7)
RETURNING id;

-- name: GetRefreshSessionWithContext :one
SELECT
    rs.id,
    rs.user_id,
    rs.membership_id,
    rs.organization_id,
    rs.family_id,
    rs.auth_version,
    rs.expires_at,
    rs.revoked_at,
    rs.replaced_by_token_hash
FROM refresh_sessions rs
JOIN organizations o ON o.id = rs.organization_id
JOIN organization_memberships m ON m.id = rs.membership_id
JOIN users u ON u.id = rs.user_id
WHERE rs.token_hash = $1
  AND o.status = 'ACTIVE'
  AND m.status = 'ACTIVE'
  AND u.status = 'ACTIVE'
  AND u.auth_version = rs.auth_version
FOR UPDATE;

-- name: FindRefreshSessionByTokenHash :one
SELECT
    id,
    user_id,
    membership_id,
    organization_id,
    family_id,
    auth_version,
    expires_at,
    revoked_at,
    replaced_by_token_hash
FROM refresh_sessions
WHERE token_hash = $1
LIMIT 1;

-- name: MarkSessionReplaced :exec
UPDATE refresh_sessions
SET replaced_by_token_hash = $2
WHERE id = $1;

-- name: RevokeSession :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: RevokeFamily :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE family_id = $1
  AND revoked_at IS NULL;

-- name: RevokeUserSessions :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: GetRolesByMembershipID :many
SELECT role
FROM membership_roles
WHERE membership_id = $1
ORDER BY role;

-- name: BumpUserAuthVersion :exec
UPDATE users
SET auth_version = auth_version + 1,
    must_change_password = false
WHERE id = $1;

-- name: UpdateLoginPassword :execrows
UPDATE membership_login_names
SET password_hash = $1
WHERE user_id = $2
  AND organization_id = $3;

-- name: CountFailedLoginAttempts :one
SELECT COUNT(*)
FROM login_attempts
WHERE organization_id = $1
  AND username_normalized = lower($2)
  AND attempted_at >= now() - $3::interval;

-- name: InsertFailedLoginAttempt :exec
INSERT INTO login_attempts (organization_id, username_normalized)
VALUES ($1, lower($2));

-- name: ClearLoginAttempts :exec
DELETE FROM login_attempts
WHERE organization_id = $1
  AND username_normalized = lower($2);

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
