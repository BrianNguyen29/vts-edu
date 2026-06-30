-- 000008_membership_roles.sql
-- Multi-role membership support and dev/E2E seed accounts.

CREATE TABLE IF NOT EXISTS membership_roles (
    membership_id uuid NOT NULL REFERENCES organization_memberships(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('student','teacher','admin')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (membership_id, role)
);

DO $$
DECLARE
    org_id uuid;
    hs_membership_id uuid;
    gv_user_id uuid;
    gv_membership_id uuid;
    admin_user_id uuid;
    admin_membership_id uuid;
    -- Hash for Password123! using the project Argon2id parameters (dev/E2E only).
    dev_password_hash text := '$argon2id$v=19$m=65536,t=3,p=4$1g6Ot1/Ps3bNRlWCAiM9mA$e194j5UoiFL4BHv+vjP4yL32dPhq6r8ybAfR4ekSsBE';
BEGIN
    SELECT id INTO org_id FROM organizations WHERE code = 'school-a';
    IF org_id IS NULL THEN
        RETURN;
    END IF;

    SELECT m.id INTO hs_membership_id
    FROM organization_memberships m
    JOIN membership_login_names ln
        ON ln.organization_id = m.organization_id AND ln.user_id = m.user_id
    WHERE m.organization_id = org_id
      AND lower(ln.username_normalized) = 'hs001'
    LIMIT 1;

    IF hs_membership_id IS NOT NULL THEN
        INSERT INTO membership_roles (membership_id, role)
        VALUES (hs_membership_id, 'student')
        ON CONFLICT DO NOTHING;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM membership_login_names
        WHERE organization_id = org_id AND lower(username_normalized) = 'gv001'
    ) THEN
        INSERT INTO users DEFAULT VALUES RETURNING id INTO gv_user_id;
        INSERT INTO organization_memberships (organization_id, user_id)
        VALUES (org_id, gv_user_id)
        RETURNING id INTO gv_membership_id;
        INSERT INTO membership_login_names (
            organization_id, username_normalized, user_id, password_hash
        ) VALUES (org_id, 'gv001', gv_user_id, dev_password_hash);
        INSERT INTO membership_roles (membership_id, role)
        VALUES (gv_membership_id, 'teacher');
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM membership_login_names
        WHERE organization_id = org_id AND lower(username_normalized) = 'admin001'
    ) THEN
        INSERT INTO users DEFAULT VALUES RETURNING id INTO admin_user_id;
        INSERT INTO organization_memberships (organization_id, user_id)
        VALUES (org_id, admin_user_id)
        RETURNING id INTO admin_membership_id;
        INSERT INTO membership_login_names (
            organization_id, username_normalized, user_id, password_hash
        ) VALUES (org_id, 'admin001', admin_user_id, dev_password_hash);
        INSERT INTO membership_roles (membership_id, role)
        VALUES (admin_membership_id, 'admin');
    END IF;
END $$;
