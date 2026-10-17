package jobconfig

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

// Save writes jobs config to path (typically jobs/config.yaml).
func Save(path string, cfg *Config) error {
	if cfg == nil {
		return os.ErrInvalid
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// SetSourceEnabled updates the yaml-backed toggle for a catalog source id.
func SetSourceEnabled(cfg *Config, sourceID string, enabled bool) bool {
	if cfg == nil {
		return false
	}
	switch sourceID {
	case "greenhouse":
		cfg.Sources.Greenhouse.Enabled = enabled
	case "remotive":
		cfg.Sources.Remotive.Enabled = enabled
	case "remoteok":
		cfg.Sources.RemoteOK.Enabled = enabled
	case "jobicy":
		cfg.Sources.Jobicy.Enabled = enabled
	case "weworkremotely":
		cfg.Sources.WeWorkRemotely.Enabled = enabled
	case "workingnomads":
		cfg.Sources.WorkingNomads.Enabled = enabled
	case "fourdayweek":
		cfg.Sources.FourDayWeek.Enabled = enabled
	default:
		return false
	}
	return true
}
