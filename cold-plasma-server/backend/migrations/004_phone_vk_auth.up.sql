ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS vk_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS auth_provider TEXT NOT NULL DEFAULT 'password';

ALTER TABLE users
    ALTER COLUMN email SET DEFAULT '',
    ALTER COLUMN phone SET DEFAULT '';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_key;

UPDATE users
SET email = ''
WHERE email IS NULL;

UPDATE users
SET phone = ''
WHERE phone IS NULL;

UPDATE users
SET phone = CASE
    WHEN length(regexp_replace(phone, '\D', '', 'g')) = 11 AND left(regexp_replace(phone, '\D', '', 'g'), 1) = '8'
        THEN '+7' || right(regexp_replace(phone, '\D', '', 'g'), 10)
    WHEN length(regexp_replace(phone, '\D', '', 'g')) = 11 AND left(regexp_replace(phone, '\D', '', 'g'), 1) = '7'
        THEN '+' || regexp_replace(phone, '\D', '', 'g')
    WHEN length(regexp_replace(phone, '\D', '', 'g')) = 10 AND left(regexp_replace(phone, '\D', '', 'g'), 1) = '9'
        THEN '+7' || regexp_replace(phone, '\D', '', 'g')
    ELSE ''
END
WHERE phone <> '';

UPDATE users
SET auth_provider = 'password'
WHERE auth_provider = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_not_empty ON users (lower(email))
WHERE email <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_not_empty ON users (phone)
WHERE phone <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_vk_id_not_empty ON users (vk_id)
WHERE vk_id <> '';

CREATE TABLE IF NOT EXISTS phone_verification_codes (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone          TEXT        NOT NULL,
    code_hash      TEXT        NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL,
    attempts_left  INT         NOT NULL DEFAULT 5,
    consumed_at    TIMESTAMPTZ NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_phone_verification_codes_user_active
    ON phone_verification_codes (user_id, created_at DESC)
    WHERE consumed_at IS NULL;
