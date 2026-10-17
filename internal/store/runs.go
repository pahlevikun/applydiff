package store

import (
	"context"
	"database/sql"
	"time"
)

// AsyncRun is a persisted background job.
type AsyncRun struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	JobKey     string    `json:"job_key"`
	Status     string    `json:"status"`
	ResultJSON string    `json:"result_json,omitempty"`
	Error      string    `json:"error,omitempty"`
	Created    time.Time `json:"created_at"`
	Updated    time.Time `json:"updated_at"`
}

const (
	RunStatusRunning = "running"
	RunStatusDone    = "done"
	RunStatusFailed  = "failed"
)

func (s *Store) InsertRun(ctx context.Context, r AsyncRun) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO async_runs (id, kind, job_key, status, result_json, error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Kind, r.JobKey, r.Status, r.ResultJSON, r.Error,
		r.Created.UTC().Format(time.RFC3339), r.Updated.UTC().Format(time.RFC3339))
	return err
}

func (s *Store) UpdateRun(ctx context.Context, r AsyncRun) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE async_runs SET status = ?, result_json = ?, error = ?, updated_at = ?
		WHERE id = ?`,
		r.Status, r.ResultJSON, r.Error, r.Updated.UTC().Format(time.RFC3339), r.ID)
	return err
}

func (s *Store) ListRecentRuns(ctx context.Context, limit int) ([]AsyncRun, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, kind, job_key, status, result_json, error, created_at, updated_at
		FROM async_runs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AsyncRun
	for rows.Next() {
		var r AsyncRun
		var created, updated string
		if err := rows.Scan(&r.ID, &r.Kind, &r.JobKey, &r.Status, &r.ResultJSON, &r.Error, &created, &updated); err != nil {
			return nil, err
		}
		r.Created, _ = time.Parse(time.RFC3339, created)
		r.Updated, _ = time.Parse(time.RFC3339, updated)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetRun(ctx context.Context, id string) (AsyncRun, error) {
	var r AsyncRun
	var created, updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, kind, job_key, status, result_json, error, created_at, updated_at
		FROM async_runs WHERE id = ?`, id).Scan(
		&r.ID, &r.Kind, &r.JobKey, &r.Status, &r.ResultJSON, &r.Error, &created, &updated)
	if err != nil {
		return AsyncRun{}, err
	}
	r.Created, _ = time.Parse(time.RFC3339, created)
	r.Updated, _ = time.Parse(time.RFC3339, updated)
	return r, nil
}

// FailStaleRuns marks interrupted runs as failed on startup.
func (s *Store) FailStaleRuns(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE async_runs SET status = ?, error = ?, updated_at = datetime('now')
		WHERE status = ?`, RunStatusFailed, "interrupted", RunStatusRunning)
	return err
}

func (s *Store) LogAction(ctx context.Context, jobID, action, actor, detail string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO action_log (job_id, action, actor, detail) VALUES (?, ?, ?, ?)`,
		jobID, action, actor, detail)
	return err
}

func (s *Store) ListActions(ctx context.Context, jobID string) ([]ActionLogEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT job_id, action, actor, detail, created_at FROM action_log
		WHERE job_id = ? ORDER BY id DESC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ActionLogEntry
	for rows.Next() {
		var e ActionLogEntry
		if err := rows.Scan(&e.JobID, &e.Action, &e.Actor, &e.Detail, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ActionLogEntry is one audit row.
type ActionLogEntry struct {
	JobID     string `json:"job_id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Detail    string `json:"detail,omitempty"`
	CreatedAt string `json:"created_at"`
}

// TableMissing returns true when migration has not been applied yet.
func (s *Store) TableMissing(ctx context.Context, table string) bool {
	var name string
	err := s.db.QueryRowContext(ctx, `
		SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
	return err == sql.ErrNoRows
}
