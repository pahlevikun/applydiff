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
