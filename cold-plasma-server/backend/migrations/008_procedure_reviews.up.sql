ALTER TABLE procedures
    ADD COLUMN IF NOT EXISTS video_url TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS procedure_reviews (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    procedure_id BIGINT NOT NULL REFERENCES procedures(id) ON DELETE CASCADE,
    rating       INT NOT NULL CHECK (rating >= 0 AND rating <= 5),
    text         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, procedure_id)
);

CREATE INDEX IF NOT EXISTS idx_procedure_reviews_procedure_id
    ON procedure_reviews(procedure_id, created_at DESC);
