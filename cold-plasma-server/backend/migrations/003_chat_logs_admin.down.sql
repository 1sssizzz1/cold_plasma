DROP INDEX IF EXISTS idx_chat_logs_created_at;

ALTER TABLE chat_logs
    DROP COLUMN IF EXISTS intent,
    DROP COLUMN IF EXISTS user_phone,
    DROP COLUMN IF EXISTS user_email,
    DROP COLUMN IF EXISTS user_name;
