ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified_at;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verification_sent_at;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified;

