package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pahlevikun/applydiff/internal/application"
)

func (s *Store) CreateApplication(ctx context.Context, jobID string) (*application.Application, error) {
	if jobID == "" {
		return nil, fmt.Errorf("store: job_id required")
	}
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO applications (job_id, status, follow_up_note, cv_snapshot, channel, referrer, notes, updated_at)
		VALUES (?, ?, '', '', 'board', '', '', ?)
		ON CONFLICT(job_id) DO NOTHING
	`, jobID, application.StatusDiscovered, now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	return s.GetApplication(ctx, jobID)
}

func (s *Store) GetApplication(ctx context.Context, jobID string) (*application.Application, error) {
	var a application.Application
	var updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT job_id, status, follow_up_note, cv_snapshot, channel, referrer, notes, updated_at
		FROM applications WHERE job_id = ?
	`, jobID).Scan(&a.JobID, &a.Status, &a.FollowUpNote, &a.CVSnapshot, &a.Channel, &a.Referrer, &a.Notes, &updated)
	if err != nil {
		return nil, err
	}
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &a, nil
}

func (s *Store) ListApplications(ctx context.Context) ([]application.ListItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT j.id, j.company, j.title, j.url,
		       COALESCE(a.status, 'discovered'),
		       COALESCE(a.follow_up_note, ''),
		       COALESCE(a.channel, 'board'),
		       COALESCE(a.referrer, ''),
		       COALESCE(a.notes, ''),
		       COALESCE(a.updated_at, '')
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		WHERE j.status != 'dead'
		ORDER BY COALESCE(a.updated_at, j.posted_at) DESC, j.title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []application.ListItem
	for rows.Next() {
		var item application.ListItem
		var updated string
		if err := rows.Scan(&item.JobID, &item.Company, &item.Title, &item.URL, &item.Status, &item.FollowUpNote, &item.Channel, &item.Referrer, &item.Notes, &updated); err != nil {
			return nil, err
		}
		if updated != "" {
			item.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) TransitionApplication(ctx context.Context, jobID, toStatus, followUpNote string) (*application.Application, error) {
	app, err := s.GetApplication(ctx, jobID)
	if err == sql.ErrNoRows {
		app, err = s.CreateApplication(ctx, jobID)
	}
	if err != nil {
		return nil, err
	}
	newStatus, err := application.Transition(app.Status, toStatus)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	note := app.FollowUpNote
	if followUpNote != "" {
		note = followUpNote
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE applications SET status = ?, follow_up_note = ?, updated_at = ? WHERE job_id = ?
	`, newStatus, note, now, jobID)
	if err != nil {
		return nil, err
	}
	return s.GetApplication(ctx, jobID)
}

func (s *Store) SetApplicationChannel(ctx context.Context, jobID, channel, referrer, notes string) error {
	if channel == "" {
		channel = "board"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE applications SET channel = ?, referrer = ?, notes = ?, updated_at = ? WHERE job_id = ?
	`, channel, referrer, notes, now, jobID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) SetApplicationCVSnapshot(ctx context.Context, jobID, cvJSON string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `
		UPDATE applications SET cv_snapshot = ?, updated_at = ? WHERE job_id = ?
	`, cvJSON, now, jobID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var v string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	return v, err
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}
