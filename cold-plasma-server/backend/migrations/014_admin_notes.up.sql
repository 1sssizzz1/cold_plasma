-- Заметки администратора занимают окна записи в календаре,
-- блокируя выбор этого времени пользователями.
CREATE TABLE IF NOT EXISTS admin_notes (
    id         BIGSERIAL PRIMARY KEY,
    start_at   TIMESTAMPTZ NOT NULL,
    end_at     TIMESTAMPTZ NOT NULL,
    title      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT admin_notes_time_order CHECK (end_at > start_at)
);

CREATE INDEX IF NOT EXISTS idx_admin_notes_start_at ON admin_notes (start_at);
