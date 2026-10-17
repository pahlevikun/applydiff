package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/match"
)

func (s *Store) SaveMatchResult(ctx context.Context, res match.Result) error {
	if res.JobID == "" {
		return fmt.Errorf("store: match job_id required")
	}
	dims, err := json.Marshal(res.Dimensions)
	if err != nil {
		return err
	}
	reqs, err := json.Marshal(res.Requirements)
	if err != nil {
		return err
	}
	leg, err := json.Marshal(res.Legitimacy)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO match_results (job_id, total, dimensions_json, requirements_json, legitimacy_json, jd_text, computed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(job_id) DO UPDATE SET
			total = excluded.total,
			dimensions_json = excluded.dimensions_json,
			requirements_json = excluded.requirements_json,
			legitimacy_json = excluded.legitimacy_json,
			jd_text = excluded.jd_text,
			computed_at = excluded.computed_at
	`, res.JobID, res.Total, string(dims), string(reqs), string(leg), res.JDText, res.ComputedAt)
	return err
}

func (s *Store) LoadMatchResult(ctx context.Context, jobID string) (*match.Result, error) {
	var total int
	var dimsJSON, reqsJSON, legJSON, jdText, computedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT total, dimensions_json, requirements_json, legitimacy_json, jd_text, computed_at
		FROM match_results WHERE job_id = ?
	`, jobID).Scan(&total, &dimsJSON, &reqsJSON, &legJSON, &jdText, &computedAt)
	if err != nil {
		return nil, err
	}
	var dims match.Dimensions
	if err := json.Unmarshal([]byte(dimsJSON), &dims); err != nil {
		return nil, err
	}
	var reqs []match.Requirement
	if err := json.Unmarshal([]byte(reqsJSON), &reqs); err != nil {
		return nil, err
	}
	var leg match.Legitimacy
	if err := json.Unmarshal([]byte(legJSON), &leg); err != nil {
		return nil, err
	}
	return &match.Result{
		JobID:        jobID,
		Total:        total,
		Dimensions:   dims,
		Requirements: reqs,
		Legitimacy:   leg,
		JDText:       jdText,
		ComputedAt:   computedAt,
	}, nil
}

func (s *Store) GetJob(ctx context.Context, id string) (*job.Job, error) {
	row := s.db.QueryRowContext(ctx, jobSelectSQL+` WHERE id = ?`, id)
	j, err := scanJobRow(row)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *Store) UpdateJobStatus(ctx context.Context, id, status string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE jobs SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
