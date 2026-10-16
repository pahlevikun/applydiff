package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func (s *Store) SaveMaster(ctx context.Context, doc *cv.Document) error {
	if doc == nil {
		return fmt.Errorf("store: master cv is nil")
	}
	if err := cv.Validate(doc); err != nil {
		return err
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO cv_master (id, person, data, updated_at)
		VALUES (1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			person = excluded.person,
			data = excluded.data,
			updated_at = excluded.updated_at
	`, doc.Person, string(data), now)
	return err
}

func (s *Store) LoadMaster(ctx context.Context) (*cv.Document, error) {
	var data string
	err := s.db.QueryRowContext(ctx, `SELECT data FROM cv_master WHERE id = 1`).Scan(&data)
	if err != nil {
		return nil, err
	}
	return cv.LoadJSON([]byte(data))
}

func (s *Store) CreateBranch(ctx context.Context, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("store: job_id required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cv_branches (job_id, hunks, created_at)
		VALUES (?, '[]', ?)
		ON CONFLICT(job_id) DO NOTHING
	`, jobID, now)
	return err
}

func (s *Store) ApplyHunk(ctx context.Context, jobID string, hunk cv.Hunk) error {
	if jobID == "" {
		return fmt.Errorf("store: job_id required")
	}
	hunks, err := s.loadBranchHunks(ctx, jobID)
	if err != nil {
		return err
	}
	replaced := false
	for i := range hunks {
		if hunks[i].FieldID == hunk.FieldID {
			hunks[i] = hunk
			replaced = true
			break
		}
	}
	if !replaced {
		hunks = append(hunks, hunk)
	}
	raw, err := json.Marshal(hunks)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE cv_branches SET hunks = ? WHERE job_id = ?`, string(raw), jobID)
	return err
}

func (s *Store) LoadBranch(ctx context.Context, jobID string) (*cv.Document, error) {
	master, err := s.LoadMaster(ctx)
	if err != nil {
		return nil, err
	}
	hunks, err := s.loadBranchHunks(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return cv.ApplyHunks(master, hunks)
}

func (s *Store) loadBranchHunks(ctx context.Context, jobID string) ([]cv.Hunk, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT hunks FROM cv_branches WHERE job_id = ?`, jobID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var hunks []cv.Hunk
	if err := json.Unmarshal([]byte(raw), &hunks); err != nil {
		return nil, err
	}
	return hunks, nil
}
