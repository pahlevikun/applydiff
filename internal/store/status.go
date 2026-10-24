package store

import (
	"context"
	"database/sql"
	"time"
)

// SystemStatus exposes daemon and scan timestamps for the web UI.
type SystemStatus struct {
	LastScanAt   *time.Time
	DaemonSeenAt *time.Time
}

// LoadSystemStatus returns the latest scan time across sources and the latest daemon tick.
func (s *Store) LoadSystemStatus(ctx context.Context) (SystemStatus, error) {
	out := SystemStatus{}
	scans, err := s.LoadSourceScanResults(ctx)
	if err != nil {
		return out, err
	}
	for _, row := range scans {
		if row.At == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, row.At)
		if err != nil {
			continue
		}
		if out.LastScanAt == nil || t.After(*out.LastScanAt) {
			out.LastScanAt = &t
		}
	}
	var finished string
	err = s.db.QueryRowContext(ctx, `
		SELECT finished_at FROM runs
		WHERE kind = 'daemon_tick' AND finished_at != ''
		ORDER BY finished_at DESC LIMIT 1
	`).Scan(&finished)
	if err != nil && err != sql.ErrNoRows {
		return out, err
	}
	if finished != "" {
		if t, err := time.Parse(time.RFC3339, finished); err == nil {
			out.DaemonSeenAt = &t
		}
	}
	if out.LastScanAt == nil {
		var scanFinished string
		err = s.db.QueryRowContext(ctx, `
			SELECT finished_at FROM runs
			WHERE kind IN ('scan', 'daemon_tick') AND finished_at != ''
			ORDER BY finished_at DESC LIMIT 1
		`).Scan(&scanFinished)
		if err == nil && scanFinished != "" {
			if t, err := time.Parse(time.RFC3339, scanFinished); err == nil {
				out.LastScanAt = &t
			}
		}
	}
	return out, nil
}
