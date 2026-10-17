package jobconfig

import (
	"context"
	"encoding/json"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/store"
)

// SourceView is one row in the sources settings grid.
type SourceView struct {
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Hint     string                  `json:"hint"`
	Enabled  bool                    `json:"enabled"`
	Scan     SourceScanMode          `json:"scan"`
	Origins  string                  `json:"origins,omitempty"`
	LastScan *store.SourceScanResult `json:"last_scan,omitempty"`
}

// ConfigView is exposed to the Sources UI.
type ConfigView struct {
	Sources    []SourceView `json:"sources"`
	Queries    []string     `json:"queries"`
	Filters    filter.Spec  `json:"filters"`
	LLM        LLMView      `json:"llm"`
	ConfigPath string       `json:"config_path,omitempty"`
}

type LLMView struct {
	Mode  string `json:"mode"`
	Model string `json:"model"`
}

func catalogSourceViews() []SourceView {
	cat := SourceCatalog()
	out := make([]SourceView, len(cat))
	for i, e := range cat {
		out[i] = SourceView{
			ID:      e.ID,
			Name:    e.Name,
			Hint:    e.Hint,
			Enabled: e.DefaultOn,
			Scan:    e.Scan,
			Origins: e.Origins,
		}
	}
	return out
}

// BuildView merges yaml config, persisted toggles, and LLM env.
func (c *Config) BuildView(ctx context.Context, loadOverride func(ctx context.Context, key string) (string, error), lastScan map[string]store.SourceScanResult) ConfigView {
	sources := catalogSourceViews()
	for i := range sources {
		for _, e := range SourceCatalog() {
			if e.ID != sources[i].ID {
				continue
			}
			if e.ConfigEnabled != nil {
				sources[i].Enabled = e.ConfigEnabled(c)
			}
			break
		}
	}

	if loadOverride != nil {
		raw, err := loadOverride(ctx, "source_enabled")
		if err == nil && raw != "" {
			var overrides map[string]bool
			if json.Unmarshal([]byte(raw), &overrides) == nil {
				for i := range sources {
					if v, ok := overrides[sources[i].ID]; ok {
						sources[i].Enabled = v
					}
				}
			}
		}
	}

	for i := range sources {
		if res, ok := lastScan[sources[i].ID]; ok {
			cp := res
			sources[i].LastScan = &cp
		}
	}

	llmCfg := llm.ConfigFromEnv()
	return ConfigView{
		Sources: sources,
		Queries: c.Queries,
		Filters: c.Filters.EffectiveSpec(),
		LLM: LLMView{
			Mode:  llmCfg.Provider,
			Model: llmCfg.Model,
		},
	}
}
