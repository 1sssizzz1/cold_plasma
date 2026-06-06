ALTER TABLE users
    ADD COLUMN IF NOT EXISTS telegram_chat_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS telegram_username TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS telegram_linked_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS telegram_link_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_telegram_link_tokens_user_active
    ON telegram_link_tokens(user_id, created_at DESC)
    WHERE consumed_at IS NULL;
