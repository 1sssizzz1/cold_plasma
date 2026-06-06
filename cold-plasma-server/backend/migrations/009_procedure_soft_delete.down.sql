DROP INDEX IF EXISTS idx_procedures_not_deleted;

ALTER TABLE procedures
    DROP COLUMN IF EXISTS deleted_at;
