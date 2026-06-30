-- 000012_admin_user_management.sql
-- Add minimal user profile fields needed for admin user management.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email text,
ADD COLUMN IF NOT EXISTS display_name text;

UPDATE users
SET display_name = ln.username_normalized,
    email = ln.username_normalized || '@vts-edu.local'
FROM membership_login_names ln
WHERE users.id = ln.user_id
  AND users.display_name IS NULL;
