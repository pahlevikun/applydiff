package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PipelineRun is one daemon tick (or other scheduled engine run).
type PipelineRun struct {
	ID         string
	Kind       string
	StartedAt  time.Time
	FinishedAt time.Time
	Fetched    int
	Stored     int
	Skipped    int
	Error      string
	DetailJSON string
}

// InsertPipelineRun persists a completed run row.
// CountPipelineRuns counts rows for a kind (empty kind = all).
func (s *Store) CountPipelineRuns(ctx context.Context, kind string) (int, error) {
	var n int
	if kind == "" {
		err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs`).Scan(&n)
		return n, err
	}
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs WHERE kind = ?`, kind).Scan(&n)
	return n, err
}

func (s *Store) InsertPipelineRun(ctx context.Context, run PipelineRun) error {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO runs (id, kind, started_at, finished_at, fetched, stored, skipped, error, detail_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, run.ID, run.Kind, run.StartedAt.UTC().Format(time.RFC3339), run.FinishedAt.UTC().Format(time.RFC3339),
		run.Fetched, run.Stored, run.Skipped, run.Error, run.DetailJSON)
	return err
}

// CountProposalsSince counts tailor proposals created at or after since.
func (s *Store) CountProposalsSince(ctx context.Context, since time.Time) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tailor_proposals WHERE created_at >= ?
	`, since.UTC().Format(time.RFC3339)).Scan(&n)
	return n, err
}

// ListJobsNeedingMatch returns active feed jobs without a stored match result.
func (s *Store) ListJobsNeedingMatch(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT j.id FROM jobs j
		LEFT JOIN match_results m ON m.job_id = j.id
		WHERE j.status NOT IN ('dead', 'skipped', 'discarded')
		  AND (j.filter_skip_reason = '' OR j.filter_skip_reason IS NULL)
		  AND m.job_id IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// InboxJob is job metadata joined for inbox classification.
type InboxJob struct {
	ID               string
	Company          string
	Title            string
	Status           string
	FilterSkipReason string
	UpdatedAt        time.Time
}

// InboxRow is joined job + pipeline state for the derived inbox.
type InboxRow struct {
	Job          InboxJob
	AppStatus    string
	AppUpdatedAt time.Time
	MatchTotal   int
	HasProposal  bool
	PendingHunks int
}

// ListInboxRows loads jobs with application, match, and proposal summary.
// Materialize queue_items only if this query becomes slow (see internal/inbox).
func (s *Store) ListInboxRows(ctx context.Context) ([]InboxRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT j.id, j.company, j.title, j.status, j.filter_skip_reason, j.posted_at,
		       COALESCE(a.status, ''), COALESCE(a.updated_at, ''),
		       COALESCE(m.total, -1),
		       EXISTS(SELECT 1 FROM tailor_proposals p WHERE p.job_id = j.id),
		       COALESCE((
		         SELECT COUNT(*) FROM proposal_hunks h
		         JOIN tailor_proposals p ON p.id = h.proposal_id
		         WHERE p.job_id = j.id AND h.status = 'pending'
		       ), 0)
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		LEFT JOIN match_results m ON m.job_id = j.id
		WHERE j.status NOT IN ('dead', 'discarded')
		ORDER BY COALESCE(a.updated_at, j.posted_at) DESC, j.title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InboxRow
	for rows.Next() {
		var r InboxRow
		var posted, appUpdated string
		if err := rows.Scan(
			&r.Job.ID, &r.Job.Company, &r.Job.Title, &r.Job.Status, &r.Job.FilterSkipReason, &posted,
			&r.AppStatus, &appUpdated,
			&r.MatchTotal, &r.HasProposal, &r.PendingHunks,
		); err != nil {
			return nil, err
		}
		if appUpdated != "" {
			r.AppUpdatedAt, _ = time.Parse(time.RFC3339, appUpdated)
		}
		if r.AppUpdatedAt.IsZero() && posted != "" {
			r.Job.UpdatedAt, _ = time.Parse(time.RFC3339, posted)
		} else {
			r.Job.UpdatedAt = r.AppUpdatedAt
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PipelineRunDetail is stored in runs.detail_json for daemon ticks.
type PipelineRunDetail struct {
	Sources       map[string]SourceTickResult `json:"sources,omitempty"`
	Matched       int                       `json:"matched,omitempty"`
	Drafted       int                       `json:"drafted,omitempty"`
	DraftNote     string                    `json:"draft_note,omitempty"`
	DryRun        bool                      `json:"dry_run,omitempty"`
	Notify        bool                      `json:"notify,omitempty"`
}

type SourceTickResult struct {
	Fetched int    `json:"fetched"`
	Stored  int    `json:"stored"`
	Skipped int    `json:"skipped"`
	Error   string `json:"error,omitempty"`
}

func MarshalPipelineDetail(d PipelineRunDetail) string {
	raw, _ := json.Marshal(d)
	return string(raw)
}
