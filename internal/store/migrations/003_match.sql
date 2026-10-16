ALTER TABLE jobs ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN salary_post TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN salary_ats TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS match_results (
    job_id TEXT PRIMARY KEY,
    total INTEGER NOT NULL,
    dimensions_json TEXT NOT NULL,
    requirements_json TEXT NOT NULL,
    legitimacy_json TEXT NOT NULL,
    jd_text TEXT NOT NULL DEFAULT '',
    computed_at TEXT NOT NULL,
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);
