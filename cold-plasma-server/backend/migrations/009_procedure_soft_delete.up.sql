ALTER TABLE procedures
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_procedures_not_deleted
    ON procedures(id)
    WHERE deleted_at IS NULL;
