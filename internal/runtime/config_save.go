package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
)

// ConfigPatch is a partial update from the Sources UI.
type ConfigPatch struct {
	SourceID *string
	Enabled  *bool
	Queries  *[]string
	Filters  *filter.Spec
}

// ApplyConfigPatch updates in-memory config, persists to config.yaml, and refreshes scan sources.
func (a *App) ApplyConfigPatch(ctx context.Context, patch ConfigPatch) error {
	if a == nil || a.Config == nil {
		return fmt.Errorf("config not loaded")
	}
	if patch.SourceID != nil && *patch.SourceID != "" {
		if patch.Enabled == nil {
			return fmt.Errorf("enabled required with source_id")
		}
		if err := a.setSourceOverride(ctx, *patch.SourceID, *patch.Enabled); err != nil {
			return err
		}
	}
	if patch.Queries != nil {
		a.Config.Queries = append([]string(nil), (*patch.Queries)...)
	}
	if patch.Filters != nil {
		a.Config.Filters.Spec = patch.Filters.Clone()
	}
	if a.CfgPath == "" {
		return fmt.Errorf("config path not set")
	}
	if err := jobconfig.Save(a.CfgPath, a.Config); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	a.reloadScanSources()
	return nil
}

func (a *App) setSourceOverride(ctx context.Context, sourceID string, enabled bool) error {
	view := a.SourcesConfigView(ctx)
	overrides := map[string]bool{}
	for _, src := range view.Sources {
		overrides[src.ID] = src.Enabled
	}
	overrides[sourceID] = enabled
	raw, err := json.Marshal(overrides)
	if err != nil {
		return err
	}
	if err := a.Store.SetSetting(ctx, "source_enabled", string(raw)); err != nil {
		return err
	}
	if !jobconfig.SetSourceEnabled(a.Config, sourceID, enabled) {
		return fmt.Errorf("unknown source_id %q", sourceID)
	}
	return nil
}

func (a *App) reloadScanSources() {
	registerSources(a.Config)
	if a.Runner != nil {
		a.Runner.Config = a.Config
	}
}
