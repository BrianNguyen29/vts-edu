-- 000007_attempt_grading.sql
-- Minimal MCQ grading support for the demo attempt.

ALTER TABLE attempt_items
ADD COLUMN IF NOT EXISTS answer_key_json jsonb NOT NULL DEFAULT '{}';

ALTER TABLE attempts
ADD COLUMN IF NOT EXISTS score numeric(10,2),
ADD COLUMN IF NOT EXISTS max_score numeric(10,2),
ADD COLUMN IF NOT EXISTS grading_status text;

UPDATE attempt_items
SET answer_key_json = '{"correct_option":"A"}'
WHERE attempt_id = '00000000-0000-4000-8000-000000000001'::uuid
  AND position = 1;

UPDATE attempt_items
SET answer_key_json = '{"correct_option":"B"}'
WHERE attempt_id = '00000000-0000-4000-8000-000000000001'::uuid
  AND position = 2;
