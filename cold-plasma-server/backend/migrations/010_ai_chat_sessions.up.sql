CREATE TABLE IF NOT EXISTS ai_chat_sessions (
    id         TEXT PRIMARY KEY,
    user_id    BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    user_name  TEXT NOT NULL DEFAULT '',
    user_email TEXT NOT NULL DEFAULT '',
    user_phone TEXT NOT NULL DEFAULT '',
    opened_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at  TIMESTAMPTZ
);

ALTER TABLE chat_logs
    ADD COLUMN IF NOT EXISTS session_id TEXT NULL REFERENCES ai_chat_sessions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_chat_logs_session_id ON chat_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_opened_at ON ai_chat_sessions(opened_at DESC);
