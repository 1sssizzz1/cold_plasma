DROP INDEX IF EXISTS idx_ai_chat_sessions_opened_at;
DROP INDEX IF EXISTS idx_chat_logs_session_id;

ALTER TABLE chat_logs
    DROP COLUMN IF EXISTS session_id;

DROP TABLE IF EXISTS ai_chat_sessions;
