ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS telegram_admin_reminder_message_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS telegram_admin_reminder_delete_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS telegram_admin_reminder_deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_bookings_telegram_admin_reminder_delete_due
    ON bookings (telegram_admin_reminder_delete_at)
    WHERE telegram_admin_reminder_delete_at IS NOT NULL
      AND telegram_admin_reminder_deleted_at IS NULL;
