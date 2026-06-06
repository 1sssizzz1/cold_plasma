CREATE INDEX IF NOT EXISTS idx_chat_logs_created_at ON chat_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_closed_at ON ai_chat_sessions(closed_at);
