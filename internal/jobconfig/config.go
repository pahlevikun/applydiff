package jobconfig

import (
	"os"

	"github.com/pahlevikun/applydiff/internal/filter"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources SourcesConfig `yaml:"sources"`
	Queries []string      `yaml:"queries"`
	Filters Filters       `yaml:"filters"`
	Caps    Caps          `yaml:"caps"`
}

type SourcesConfig struct {
	Greenhouse      GreenhouseConfig      `yaml:"greenhouse"`
	Remotive        RemotiveConfig        `yaml:"remotive"`
	RemoteOK        RemoteOKConfig        `yaml:"remoteok"`
	Jobicy          JobicyConfig          `yaml:"jobicy"`
	WeWorkRemotely  WeWorkRemotelyConfig  `yaml:"weworkremotely"`
	WorkingNomads   WorkingNomadsConfig   `yaml:"workingnomads"`
}

// SourceToggle is shared yaml for simple JSON/RSS boards.
type SourceToggle struct {
	Enabled bool     `yaml:"enabled"`
	Queries []string `yaml:"queries,omitempty"`
}

type GreenhouseConfig struct {
	Enabled bool          `yaml:"enabled"`
	Boards  []BoardConfig `yaml:"boards"`
	Queries []string      `yaml:"queries,omitempty"`
}

type BoardConfig struct {
	Token   string `yaml:"token"`
	Company string `yaml:"company"`
}

type RemotiveConfig struct {
	SourceToggle `yaml:",inline"`
}

type RemoteOKConfig struct {
	SourceToggle `yaml:",inline"`
}

type JobicyConfig struct {
	SourceToggle `yaml:",inline"`
}

type WeWorkRemotelyConfig struct {
	SourceToggle `yaml:",inline"`
	FeedURL      string `yaml:"feed_url,omitempty"`
}

type WorkingNomadsConfig struct {
	SourceToggle `yaml:",inline"`
}

// Filters is the inline filter spec from config yaml.
type Filters struct {
	Spec filter.Spec `yaml:",inline"`
}

type Caps struct {
	PerSource      int `yaml:"per_source"`
	MaxJobsPerScan int `yaml:"max_jobs_per_scan"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.Filters.MergeLegacy()
	return &cfg, nil
}

// EffectiveSpec returns the filter spec.
func (f Filters) EffectiveSpec() filter.Spec {
	return f.Spec.Clone()
}

// MergeLegacy is a no-op; legacy flat keys are represented inside Spec.
func (f *Filters) MergeLegacy() {}

// QueriesForSource returns per-source query overrides or the shared list.
func (c *Config) QueriesForSource(sourceID string, override []string) []string {
	if len(override) > 0 {
		return override
	}
	if c == nil {
		return nil
	}
	return c.Queries
}
