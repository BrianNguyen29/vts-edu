-- 000006_attempt_runtime_seeds.sql
-- Prepares the demo attempt seeded in 000004 for runtime endpoint testing.

UPDATE attempts
SET status = 'IN_PROGRESS',
    started_at = now(),
    expires_at = now() + interval '1 hour'
WHERE id = '00000000-0000-4000-8000-000000000001'::uuid;
