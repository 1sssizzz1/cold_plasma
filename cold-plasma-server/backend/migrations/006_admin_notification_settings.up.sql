CREATE TABLE IF NOT EXISTS app_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_settings (key, value)
VALUES
    ('admin_notify_telegram', 'true'),
    ('admin_notify_sms', 'false'),
    ('admin_sms_phone', ''),
    ('master_profile', '')
ON CONFLICT (key) DO NOTHING;
