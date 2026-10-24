ALTER TABLE jobs ADD COLUMN first_seen TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN role_key TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_jobs_role_key ON jobs(role_key);

UPDATE jobs
SET first_seen = CASE
    WHEN TRIM(first_seen) = '' THEN CASE
        WHEN TRIM(posted_at) != '' THEN posted_at
        ELSE datetime('now')
    END
    ELSE first_seen
END;
