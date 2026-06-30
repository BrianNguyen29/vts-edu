CREATE TABLE IF NOT EXISTS password_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_history_user_created
    ON password_history(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS login_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username_normalized text NOT NULL,
    attempted_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_org_user_attempted
    ON login_attempts(organization_id, username_normalized, attempted_at DESC);
