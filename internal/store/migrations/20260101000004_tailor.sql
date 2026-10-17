CREATE TABLE IF NOT EXISTS cv_facts (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    data TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tailor_proposals (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS proposal_hunks (
    id TEXT PRIMARY KEY,
    proposal_id TEXT NOT NULL,
    field_id TEXT NOT NULL,
    section TEXT NOT NULL,
    before_text TEXT NOT NULL,
    after_text TEXT NOT NULL,
    rationale TEXT NOT NULL,
    evidence TEXT NOT NULL DEFAULT '[]',
    tags TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'pending',
    FOREIGN KEY (proposal_id) REFERENCES tailor_proposals(id)
);

CREATE INDEX IF NOT EXISTS idx_proposal_hunks_proposal ON proposal_hunks(proposal_id);
