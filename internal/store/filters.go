package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pahlevikun/applydiff/internal/filter"
)

const activeFilterPresetKey = "active_filter_preset_id"

type FilterPreset struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Spec      filter.Spec  `json:"spec"`
	CreatedAt string       `json:"created_at"`
}

func (s *Store) ListFilterPresets(ctx context.Context) ([]FilterPreset, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, spec_json, created_at FROM filter_presets ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FilterPreset
	for rows.Next() {
		var p FilterPreset
		var specJSON string
		if err := rows.Scan(&p.ID, &p.Name, &specJSON, &p.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(specJSON), &p.Spec); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) SaveFilterPreset(ctx context.Context, name string, spec filter.Spec) (FilterPreset, error) {
	if name == "" {
		return FilterPreset{}, fmt.Errorf("preset name required")
	}
	id := uuid.NewString()
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return FilterPreset{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO filter_presets (id, name, spec_json, created_at) VALUES (?, ?, ?, ?)
	`, id, name, string(specJSON), now)
	if err != nil {
		return FilterPreset{}, err
	}
	return FilterPreset{ID: id, Name: name, Spec: spec, CreatedAt: now}, nil
}

func (s *Store) DeleteFilterPreset(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM filter_presets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	active, _ := s.GetActiveFilterPresetID(ctx)
	if active == id {
		_ = s.SetSetting(ctx, activeFilterPresetKey, "")
	}
	return nil
}

func (s *Store) GetFilterPreset(ctx context.Context, id string) (*FilterPreset, error) {
	var p FilterPreset
	var specJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, spec_json, created_at FROM filter_presets WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &specJSON, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(specJSON), &p.Spec); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetActiveFilterPresetID(ctx context.Context) (string, error) {
	return s.GetSetting(ctx, activeFilterPresetKey)
}

func (s *Store) SetActiveFilterPresetID(ctx context.Context, id string) error {
	if id != "" {
		if _, err := s.GetFilterPreset(ctx, id); err != nil {
			return err
		}
	}
	return s.SetSetting(ctx, activeFilterPresetKey, id)
}

func (s *Store) SaveFilterSpecOverride(ctx context.Context, spec filter.Spec) error {
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return s.SetSetting(ctx, "filter_spec_override", string(raw))
}

func (s *Store) LoadFilterSpecOverride(ctx context.Context) (*filter.Spec, error) {
	raw, err := s.GetSetting(ctx, "filter_spec_override")
	if err != nil || raw == "" {
		return nil, err
	}
	var spec filter.Spec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}
