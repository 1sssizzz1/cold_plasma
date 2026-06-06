DROP INDEX IF EXISTS idx_bookings_telegram_admin_reminder_delete_due;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS telegram_admin_reminder_deleted_at,
    DROP COLUMN IF EXISTS telegram_admin_reminder_delete_at,
    DROP COLUMN IF EXISTS telegram_admin_reminder_message_ids;
