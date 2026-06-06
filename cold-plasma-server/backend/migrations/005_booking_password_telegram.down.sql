DROP INDEX IF EXISTS idx_bookings_telegram_reminders;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_active;
DROP TABLE IF EXISTS password_reset_tokens;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS requested_datetimes,
    DROP COLUMN IF EXISTS notify_sms,
    DROP COLUMN IF EXISTS notify_telegram,
    DROP COLUMN IF EXISTS telegram_created_notified_at,
    DROP COLUMN IF EXISTS telegram_reminder_sent_at;
