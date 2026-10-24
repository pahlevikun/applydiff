package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/diff"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// Proposal bundles a job's tailoring proposal and hunks.
type Proposal struct {
	ID        string        `json:"id"`
	JobID     string        `json:"job_id"`
	CreatedAt string        `json:"created_at"`
	Hunks     []tailor.Hunk `json:"hunks"`
	BranchCV  *cv.Document  `json:"branch_cv,omitempty"`
}

func (s *Store) SaveFacts(ctx context.Context, facts *cv.Facts) error {
	if facts == nil {
		return fmt.Errorf("store: facts is nil")
	}
	data, err := json.Marshal(facts)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO cv_facts (id, data) VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET data = excluded.data
	`, string(data))
	return err
}

func (s *Store) LoadFacts(ctx context.Context) (*cv.Facts, error) {
	var data string
	err := s.db.QueryRowContext(ctx, `SELECT data FROM cv_facts WHERE id = 1`).Scan(&data)
	if err != nil {
		return nil, err
	}
	return cv.LoadFactsJSON([]byte(data))
}

func (s *Store) SaveProposal(ctx context.Context, jobID string, hunks []tailor.Hunk) (*Proposal, error) {
	if jobID == "" {
		return nil, fmt.Errorf("store: job_id required")
	}
	if err := s.CreateBranch(ctx, jobID); err != nil {
		return nil, err
	}
	proposalID := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var existingID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM tailor_proposals WHERE job_id = ?`, jobID).Scan(&existingID)
	if err == nil {
		_, _ = tx.ExecContext(ctx, `DELETE FROM proposal_hunks WHERE proposal_id = ?`, existingID)
		_, err = tx.ExecContext(ctx, `DELETE FROM tailor_proposals WHERE job_id = ?`, jobID)
		if err != nil {
			return nil, err
		}
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO tailor_proposals (id, job_id, created_at) VALUES (?, ?, ?)
	`, proposalID, jobID, now)
	if err != nil {
		return nil, err
	}
	for i := range hunks {
		if hunks[i].ID == "" {
			hunks[i].ID = uuid.NewString()
		}
		if hunks[i].Status == "" {
			hunks[i].Status = tailor.StatusPending
		}
		ev, _ := json.Marshal(hunks[i].Evidence)
		tags, _ := json.Marshal(hunks[i].Tags)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO proposal_hunks (
				id, proposal_id, field_id, section, before_text, after_text,
				rationale, evidence, tags, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, hunks[i].ID, proposalID, hunks[i].FieldID, hunks[i].Section,
			hunks[i].Before, hunks[i].After, hunks[i].Rationale, string(ev), string(tags), hunks[i].Status)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	prop, err := s.LoadProposal(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if _, err := s.ApplyHunkDecisions(ctx, jobID, prop.Hunks); err != nil {
		return nil, err
	}
	return s.LoadProposal(ctx, jobID)
}

func (s *Store) LoadProposal(ctx context.Context, jobID string) (*Proposal, error) {
	var p Proposal
	err := s.db.QueryRowContext(ctx, `
		SELECT id, job_id, created_at FROM tailor_proposals WHERE job_id = ?
	`, jobID).Scan(&p.ID, &p.JobID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	hunks, err := s.loadProposalHunks(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if _, err := s.ApplyHunkDecisions(ctx, jobID, hunks); err != nil {
		return nil, err
	}
	for i := range hunks {
		hunks[i].Index = i + 1
	}
	p.Hunks = hunks
	master, err := s.LoadMaster(ctx)
	if err != nil {
		return nil, err
	}
	branch, err := diff.ApplyAllAccepted(master, hunks)
	if err != nil {
		return nil, err
	}
	p.BranchCV = branch
	return &p, nil
}

func (s *Store) loadProposalHunks(ctx context.Context, proposalID string) ([]tailor.Hunk, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, field_id, section, before_text, after_text, rationale, evidence, tags, status
		FROM proposal_hunks WHERE proposal_id = ? ORDER BY rowid ASC
	`, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []tailor.Hunk
	for rows.Next() {
		var h tailor.Hunk
		var ev, tags string
		if err := rows.Scan(&h.ID, &h.FieldID, &h.Section, &h.Before, &h.After, &h.Rationale, &ev, &tags, &h.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(ev), &h.Evidence)
		_ = json.Unmarshal([]byte(tags), &h.Tags)
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) SetHunkStatus(ctx context.Context, hunkID, status string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE proposal_hunks SET status = ? WHERE id = ?
	`, status, hunkID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetHunk(ctx context.Context, hunkID string) (tailor.Hunk, error) {
	var h tailor.Hunk
	var ev, tags string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, field_id, section, before_text, after_text, rationale, evidence, tags, status
		FROM proposal_hunks WHERE id = ?
	`, hunkID).Scan(&h.ID, &h.FieldID, &h.Section, &h.Before, &h.After, &h.Rationale, &ev, &tags, &h.Status)
	if err != nil {
		return h, err
	}
	_ = json.Unmarshal([]byte(ev), &h.Evidence)
	_ = json.Unmarshal([]byte(tags), &h.Tags)
	return h, nil
}

func (s *Store) JobIDForHunk(ctx context.Context, hunkID string) (string, error) {
	var jobID string
	err := s.db.QueryRowContext(ctx, `
		SELECT p.job_id FROM tailor_proposals p
		JOIN proposal_hunks h ON h.proposal_id = p.id
		WHERE h.id = ?
	`, hunkID).Scan(&jobID)
	return jobID, err
}
