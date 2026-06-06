ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS requested_datetimes JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS notify_sms BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notify_telegram BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS telegram_created_notified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS telegram_reminder_sent_at TIMESTAMPTZ NULL;

UPDATE bookings
SET requested_datetimes = to_jsonb(ARRAY[datetime])
WHERE requested_datetimes = '[]'::jsonb;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_active
    ON password_reset_tokens (user_id, created_at DESC)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_bookings_telegram_reminders
    ON bookings (datetime)
    WHERE notify_telegram = TRUE AND telegram_reminder_sent_at IS NULL;
