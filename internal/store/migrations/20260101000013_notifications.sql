CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    sent_at TEXT,
    error TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_kind_run
    ON notifications(kind, json_extract(payload_json, '$.run_id'));
