CREATE TABLE IF NOT EXISTS before_after_results (
    id           BIGSERIAL PRIMARY KEY,
    procedure_id BIGINT NULL REFERENCES procedures(id) ON DELETE SET NULL,
    procedure    TEXT NOT NULL DEFAULT '',
    title        TEXT NOT NULL DEFAULT '',
    description  TEXT NOT NULL DEFAULT '',
    before_url   TEXT NOT NULL,
    after_url    TEXT NOT NULL,
    is_featured  BOOLEAN NOT NULL DEFAULT false,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_before_after_public
    ON before_after_results (is_active, sort_order, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_before_after_featured
    ON before_after_results (is_featured, is_active, sort_order, created_at DESC);
