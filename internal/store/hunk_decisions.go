package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// UpsertHunkDecision stores or replaces a decision for one hunk index (0-based).
func (s *Store) UpsertHunkDecision(ctx context.Context, applicationID string, hunkIndex int, decision, text, contentHash string) (*hunkdecision.Record, error) {
	if applicationID == "" {
		return nil, fmt.Errorf("store: application_id required")
	}
	if hunkIndex < 0 {
		return nil, fmt.Errorf("store: invalid hunk_index %d", hunkIndex)
	}
	if !hunkdecision.ValidDecision(decision) {
		return nil, fmt.Errorf("store: invalid decision %q", decision)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO hunk_decisions (application_id, hunk_index, decision, text, content_hash, decided_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(application_id, hunk_index) DO UPDATE SET
			decision = excluded.decision,
			text = excluded.text,
			content_hash = excluded.content_hash,
			decided_at = excluded.decided_at
	`, applicationID, hunkIndex, decision, text, contentHash, now)
	if err != nil {
		return nil, err
	}
	return s.GetHunkDecision(ctx, applicationID, hunkIndex)
}

// GetHunkDecision loads one decision row.
func (s *Store) GetHunkDecision(ctx context.Context, applicationID string, hunkIndex int) (*hunkdecision.Record, error) {
	var rec hunkdecision.Record
	err := s.db.QueryRowContext(ctx, `
		SELECT application_id, hunk_index, decision, text, content_hash, decided_at
		FROM hunk_decisions WHERE application_id = ? AND hunk_index = ?
	`, applicationID, hunkIndex).Scan(
		&rec.ApplicationID, &rec.HunkIndex, &rec.Decision, &rec.Text, &rec.ContentHash, &rec.DecidedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// ListHunkDecisions returns all decisions for a job/application ordered by index.
func (s *Store) ListHunkDecisions(ctx context.Context, applicationID string) ([]hunkdecision.Record, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT application_id, hunk_index, decision, text, content_hash, decided_at
		FROM hunk_decisions WHERE application_id = ? ORDER BY hunk_index ASC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []hunkdecision.Record
	for rows.Next() {
		var rec hunkdecision.Record
		if err := rows.Scan(
			&rec.ApplicationID, &rec.HunkIndex, &rec.Decision, &rec.Text, &rec.ContentHash, &rec.DecidedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// ApplyHunkDecisions merges persisted decisions onto proposal hunks and syncs proposal_hunks.status.
func (s *Store) ApplyHunkDecisions(ctx context.Context, applicationID string, hunks []tailor.Hunk) ([]bool, error) {
	records, err := s.ListHunkDecisions(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	byIndex := make(map[int]hunkdecision.Record, len(records))
	for _, rec := range records {
		byIndex[rec.HunkIndex] = rec
	}
	stale := make([]bool, len(hunks))
	for i := range hunks {
		rec, ok := byIndex[i]
		if !ok {
			continue
		}
		curHash := hunkdecision.ContentHash(hunks[i])
		isStale := rec.ContentHash != "" && rec.ContentHash != curHash
		stale[i] = isStale
		hunks[i].Decision = rec.Decision
		hunks[i].Stale = isStale
		if err := hunkdecision.ApplyToHunk(&hunks[i], rec, isStale); err != nil {
			return nil, err
		}
		if !isStale {
			if err := s.SetHunkStatus(ctx, hunks[i].ID, hunks[i].Status); err != nil && err != sql.ErrNoRows {
				return nil, err
			}
		}
	}
	return stale, nil
}
