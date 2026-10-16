package store

import (
	"context"
	"database/sql"

	"github.com/pahlevikun/applydiff/internal/job"
)

func (s *Store) UpsertJob(ctx context.Context, j job.Job, dedupeKey string) error {
	if dedupeKey == "" {
		dedupeKey = job.DedupeKey(j)
	}
	if j.ID == "" {
		j.ID = dedupeKey
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO jobs (id, company, title, url, source, posted_at, remote, region, status, dedupe_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(dedupe_key) DO UPDATE SET
			company = excluded.company,
			title = excluded.title,
			url = excluded.url,
			source = excluded.source,
			posted_at = excluded.posted_at,
			remote = excluded.remote,
			region = excluded.region,
			status = excluded.status
	`, j.ID, j.Company, j.Title, j.URL, j.Source, j.PostedAt, j.Remote, j.Region, j.Status, dedupeKey)
	return err
}

func (s *Store) ListJobs(ctx context.Context) ([]job.Job, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company, title, url, source, posted_at, remote, region, status
		FROM jobs
		WHERE status != 'dead'
		ORDER BY posted_at DESC, title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []job.Job
	for rows.Next() {
		var j job.Job
		if err := rows.Scan(&j.ID, &j.Company, &j.Title, &j.URL, &j.Source, &j.PostedAt, &j.Remote, &j.Region, &j.Status); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *Store) InsertJob(ctx context.Context, j job.Job) error {
	return s.UpsertJob(ctx, j, job.DedupeKey(j))
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
