DROP TABLE IF EXISTS telegram_link_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS telegram_linked_at,
    DROP COLUMN IF EXISTS telegram_username,
    DROP COLUMN IF EXISTS telegram_chat_id;
