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
	Greenhouse GreenhouseConfig `yaml:"greenhouse"`
}

type GreenhouseConfig struct {
	Enabled bool          `yaml:"enabled"`
	Boards  []BoardConfig `yaml:"boards"`
}

type BoardConfig struct {
	Token   string `yaml:"token"`
	Company string `yaml:"company"`
}

// Filters is the filter spec from jobs/config.yaml (inline categories).
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

// EffectiveSpec returns the filter spec with legacy yaml fields applied.
func (f Filters) EffectiveSpec() filter.Spec {
	spec := f.Spec.Clone()
	filter.NormalizeLegacy(&spec, nil, spec.RemoteOnly, "", 0)
	return spec
}

// MergeLegacy applies remote_only shorthand into arrangement when needed.
func (f *Filters) MergeLegacy() {
	filter.NormalizeLegacy(&f.Spec, nil, f.Spec.RemoteOnly, "", 0)
}
