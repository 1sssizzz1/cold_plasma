DROP TABLE IF EXISTS phone_verification_codes;

DROP INDEX IF EXISTS idx_users_vk_id_not_empty;
DROP INDEX IF EXISTS idx_users_phone_not_empty;
DROP INDEX IF EXISTS idx_users_email_not_empty;

ALTER TABLE users
    DROP COLUMN IF EXISTS phone_verified_at,
    DROP COLUMN IF EXISTS vk_id,
    DROP COLUMN IF EXISTS auth_provider;

ALTER TABLE users
    ADD CONSTRAINT users_email_key UNIQUE (email);
