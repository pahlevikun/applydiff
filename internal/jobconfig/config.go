package jobconfig

import (
	"os"

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

type Filters struct {
	Regions       []string `yaml:"regions"`
	RemoteOnly    bool     `yaml:"remote_only"`
	MinSeniority  string   `yaml:"min_seniority"`
	MaxPostedDays int      `yaml:"max_posted_days"`
}

type Caps struct {
	PerSource       int `yaml:"per_source"`
	MaxJobsPerScan  int `yaml:"max_jobs_per_scan"`
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
	return &cfg, nil
}
