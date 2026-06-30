-- 000009_user_password_flags.sql
-- Force newly provisioned service accounts to change password on first login.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS must_change_password boolean NOT NULL DEFAULT false;

UPDATE users
SET must_change_password = false
WHERE id IN (
    SELECT m.user_id
    FROM organization_memberships m
    JOIN membership_login_names ln
        ON ln.organization_id = m.organization_id AND ln.user_id = m.user_id
    WHERE lower(ln.username_normalized) = 'hs001'
);

UPDATE users
SET must_change_password = true
WHERE id IN (
    SELECT m.user_id
    FROM organization_memberships m
    JOIN membership_login_names ln
        ON ln.organization_id = m.organization_id AND ln.user_id = m.user_id
    WHERE lower(ln.username_normalized) IN ('gv001', 'admin001')
);
