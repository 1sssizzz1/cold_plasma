-- Инициализация схемы БД для проекта "Холодная плазма — Северодвинск"

-- Расширения
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT        NOT NULL UNIQUE,
    name            TEXT        NOT NULL,
    phone           TEXT        NOT NULL DEFAULT '',
    password_hash   TEXT        NOT NULL,
    photo_url       TEXT        NOT NULL DEFAULT '',
    bonus_points    INT         NOT NULL DEFAULT 0,
    phone_verified  BOOLEAN     NOT NULL DEFAULT FALSE,
    is_blocked      BOOLEAN     NOT NULL DEFAULT FALSE,
    is_admin        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Процедуры
CREATE TABLE IF NOT EXISTS procedures (
    id             BIGSERIAL PRIMARY KEY,
    title          TEXT        NOT NULL,
    description    TEXT        NOT NULL DEFAULT '',
    duration_mins  INT         NOT NULL DEFAULT 0,
    price          INT         NOT NULL DEFAULT 0,
    bonus_earned   INT         NOT NULL DEFAULT 0,
    category       TEXT        NOT NULL DEFAULT '',
    image_url      TEXT        NOT NULL DEFAULT '',
    services       TEXT        NOT NULL DEFAULT '',
    duration_str   TEXT        NOT NULL DEFAULT '',
    popular        BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Записи на процедуры
CREATE TABLE IF NOT EXISTS bookings (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    procedure_id  BIGINT      NOT NULL REFERENCES procedures(id) ON DELETE RESTRICT,
    datetime      TIMESTAMPTZ NOT NULL,
    comment       TEXT        NOT NULL DEFAULT '',
    status        TEXT        NOT NULL DEFAULT 'new', -- new|confirmed|done|canceled
    bonus_used    INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_datetime ON bookings(datetime);

-- Бонусные операции
CREATE TABLE IF NOT EXISTS bonus_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT        NOT NULL, -- earn|spend|adjust
    amount      INT         NOT NULL, -- положительное число
    comment     TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bonus_logs_user_id ON bonus_logs(user_id);

-- Логи AI-чата
CREATE TABLE IF NOT EXISTS chat_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    raw_input   TEXT        NOT NULL,
    raw_output  TEXT        NOT NULL,
    ai_model    TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_logs_user_id ON chat_logs(user_id);

-- Демо-данные (процедуры)
INSERT INTO procedures (title, description, duration_mins, price, bonus_earned, category, popular, duration_str)
VALUES
    (
        'Холодная плазма (лицо)',
        'Бережная процедура холодной плазмы для улучшения текстуры кожи и сияния. Подходит для поддерживающего ухода и мягкого обновления.',
        45,
        3500,
        350,
        'Кожа',
        TRUE,
        '45 минут'
    ),
    (
        'Холодная плазма (волосистая часть головы)',
        'Процедура холодной плазмы для кожи головы: комфортно, без агрессивных вмешательств. Помогает улучшить общее состояние кожи и ухода.',
        50,
        3800,
        380,
        'Волосы',
        TRUE,
        '50 минут'
    )
ON CONFLICT DO NOTHING;

