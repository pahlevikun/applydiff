package jobconfig

import (
	"context"
	"encoding/json"

	"github.com/pahlevikun/applydiff/internal/llm"
)

// SourceView is one row in the sources settings grid.
type SourceView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Hint    string `json:"hint"`
	Enabled bool   `json:"enabled"`
}

// ConfigView is exposed to the Sources UI.
type ConfigView struct {
	Sources []SourceView `json:"sources"`
	Queries []string     `json:"queries"`
	Filters Filters      `json:"filters"`
	LLM     LLMView      `json:"llm"`
}

type LLMView struct {
	Mode  string `json:"mode"`
	Model string `json:"model"`
}

var defaultSources = []SourceView{
	{ID: "hiring_cafe", Name: "Hiring.cafe", Hint: "agg · prefer ATS links", Enabled: true},
	{ID: "greenhouse", Name: "Greenhouse", Hint: "boards.greenhouse.io", Enabled: true},
	{ID: "lever", Name: "Lever", Hint: "jobs.lever.co", Enabled: true},
	{ID: "ashby", Name: "Ashby", Hint: "jobs.ashbyhq.com", Enabled: true},
	{ID: "workable", Name: "Workable", Hint: "apply.workable.com", Enabled: true},
	{ID: "linkedin", Name: "LinkedIn Jobs", Hint: "discovery only · cap 5/day", Enabled: true},
	{ID: "wellfound", Name: "Wellfound", Hint: "startup", Enabled: true},
	{ID: "weworkremotely", Name: "We Work Remotely", Hint: "remote board", Enabled: true},
	{ID: "remoteok", Name: "Remote OK", Hint: "remote board", Enabled: true},
	{ID: "remotive", Name: "Remotive", Hint: "remote board", Enabled: true},
	{ID: "otta", Name: "Otta", Hint: "filter remote on site", Enabled: true},
	{ID: "ycombinator", Name: "Y Combinator", Hint: "workatastartup", Enabled: true},
	{ID: "alignjobs", Name: "AlignJobs.id", Hint: "ID-focused, freemium", Enabled: false},
	{ID: "manual", Name: "Manual paste", Hint: "threads / DM links", Enabled: true},
}

// BuildView merges yaml config, persisted toggles, and LLM env.
func (c *Config) BuildView(ctx context.Context, loadOverride func(ctx context.Context, key string) (string, error)) ConfigView {
	sources := make([]SourceView, len(defaultSources))
	for i, s := range defaultSources {
		sources[i] = s
	}
	sources[1].Enabled = c.Sources.Greenhouse.Enabled

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

	llmCfg := llm.ConfigFromEnv()
	return ConfigView{
		Sources: sources,
		Queries: c.Queries,
		Filters: c.Filters,
		LLM: LLMView{
			Mode:  llmCfg.Provider,
			Model: llmCfg.Model,
		},
	}
}
