package paths

import (
	"os"
	"path/filepath"

	_ "embed"
)

//go:embed default_jobs_config.yaml
var defaultJobsConfigYAML []byte

// EnsureDefaults creates the data directory layout and writes the embedded jobs config
// when config.yaml is missing.
func EnsureDefaults(p Paths) error {
	if err := os.MkdirAll(p.Home, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.Config), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Output, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Bin, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(p.Config); os.IsNotExist(err) {
		if err := os.WriteFile(p.Config, defaultJobsConfigYAML, 0o644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// DefaultJobsConfigYAML returns the embedded default config bytes (for tests).
func DefaultJobsConfigYAML() []byte {
	return defaultJobsConfigYAML
}
