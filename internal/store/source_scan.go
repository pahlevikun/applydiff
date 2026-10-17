package store

import (
	"context"
	"encoding/json"
)

const sourceScanSettingKey = "source_last_scan"

// SourceScanResult is persisted per source after each scan attempt.
type SourceScanResult struct {
	At      string `json:"at"`
	Fetched int    `json:"fetched"`
	Stored  int    `json:"stored"`
	Skipped int    `json:"skipped"`
	Error   string `json:"error,omitempty"`
}

func (s *Store) SaveSourceScanResult(ctx context.Context, sourceID string, res SourceScanResult) error {
	all, _ := s.LoadSourceScanResults(ctx)
	if all == nil {
		all = map[string]SourceScanResult{}
	}
	all[sourceID] = res
	raw, err := json.Marshal(all)
	if err != nil {
		return err
	}
	return s.SetSetting(ctx, sourceScanSettingKey, string(raw))
}

func (s *Store) LoadSourceScanResults(ctx context.Context) (map[string]SourceScanResult, error) {
	raw, err := s.GetSetting(ctx, sourceScanSettingKey)
	if err != nil || raw == "" {
		return map[string]SourceScanResult{}, err
	}
	var out map[string]SourceScanResult
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]SourceScanResult{}, err
	}
	if out == nil {
		out = map[string]SourceScanResult{}
	}
	return out, nil
}
