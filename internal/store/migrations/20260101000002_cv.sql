CREATE TABLE IF NOT EXISTS cv_master (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    person TEXT NOT NULL,
    data TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cv_branches (
    job_id TEXT PRIMARY KEY,
    hunks TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL
);
