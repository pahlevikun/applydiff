package jobconfig

import (
	_ "embed"
	"sync"

	"gopkg.in/yaml.v3"
)

// SourceScanMode tells the UI whether scan can run today.
type SourceScanMode string

const (
	SourceScanLive    SourceScanMode = "live"
	SourceScanPlanned SourceScanMode = "planned"
	SourceScanManual  SourceScanMode = "manual"
)

// CatalogEntry is the canonical row for the Sources screen and docs.
type CatalogEntry struct {
	ID            string
	Name          string
	Hint          string
	DefaultOn     bool
	Scan          SourceScanMode
	Origins       string
	ConfigEnabled func(c *Config) bool
}

//go:embed sources_catalog.yaml
var sourcesCatalogYAML []byte

type catalogYAMLRow struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Hint      string `yaml:"hint"`
	Scan      string `yaml:"scan"`
	DefaultOn bool   `yaml:"default_on"`
	Origins   string `yaml:"origins"`
}

var (
	catalogOnce sync.Once
	catalogMemo []CatalogEntry
)

// liveConfigToggles maps catalog ids to jobs/config.yaml switches.
var liveConfigToggles = map[string]func(*Config) bool{
	"greenhouse":      func(c *Config) bool { return c != nil && c.Sources.Greenhouse.Enabled },
	"remotive":        func(c *Config) bool { return c != nil && c.Sources.Remotive.Enabled },
	"remoteok":        func(c *Config) bool { return c != nil && c.Sources.RemoteOK.Enabled },
	"jobicy":          func(c *Config) bool { return c != nil && c.Sources.Jobicy.Enabled },
	"weworkremotely":  func(c *Config) bool { return c != nil && c.Sources.WeWorkRemotely.Enabled },
	"workingnomads":   func(c *Config) bool { return c != nil && c.Sources.WorkingNomads.Enabled },
	"fourdayweek":     func(c *Config) bool { return c != nil && c.Sources.FourDayWeek.Enabled },
}

// SourceCatalog is the full inventory (career-ops + ai-job-search + job-sequencer + plugins).
func SourceCatalog() []CatalogEntry {
	catalogOnce.Do(func() {
		catalogMemo = loadSourceCatalog()
	})
	return catalogMemo
}

func loadSourceCatalog() []CatalogEntry {
	var rows []catalogYAMLRow
	if err := yaml.Unmarshal(sourcesCatalogYAML, &rows); err != nil {
		return nil
	}
	out := make([]CatalogEntry, 0, len(rows))
	for _, r := range rows {
		e := CatalogEntry{
			ID:        r.ID,
			Name:      r.Name,
			Hint:      r.Hint,
			DefaultOn: r.DefaultOn,
			Scan:      SourceScanMode(r.Scan),
			Origins:   r.Origins,
		}
		if fn, ok := liveConfigToggles[r.ID]; ok {
			e.ConfigEnabled = fn
		}
		out = append(out, e)
	}
	return out
}

// LiveSourceIDs returns ids with registered fetchers (for docs and tests).
func LiveSourceIDs() []string {
	var out []string
	for _, e := range SourceCatalog() {
		if e.Scan == SourceScanLive {
			out = append(out, e.ID)
		}
	}
	return out
}

// CatalogStats returns counts by scan mode.
func CatalogStats() (total, live, planned, manual int) {
	for _, e := range SourceCatalog() {
		total++
		switch e.Scan {
		case SourceScanLive:
			live++
		case SourceScanPlanned:
			planned++
		case SourceScanManual:
			manual++
		}
	}
	return total, live, planned, manual
}
