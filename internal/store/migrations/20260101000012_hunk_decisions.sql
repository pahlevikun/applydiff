CREATE TABLE IF NOT EXISTS hunk_decisions (
    application_id TEXT NOT NULL,
    hunk_index INTEGER NOT NULL,
    decision TEXT NOT NULL CHECK (decision IN ('accept', 'reject', 'edit', 'skip')),
    text TEXT NOT NULL DEFAULT '',
    content_hash TEXT NOT NULL DEFAULT '',
    decided_at TEXT NOT NULL,
    PRIMARY KEY (application_id, hunk_index),
    FOREIGN KEY (application_id) REFERENCES jobs(id)
);

CREATE INDEX IF NOT EXISTS idx_hunk_decisions_app ON hunk_decisions(application_id);
