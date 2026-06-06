DROP TABLE IF EXISTS procedure_reviews;

ALTER TABLE procedures
    DROP COLUMN IF EXISTS video_url;
