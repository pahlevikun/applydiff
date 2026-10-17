package runtime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/paths"
)

// DoctorReport is the structured output for applydiff doctor --json.
type DoctorReport struct {
	Home        string            `json:"home"`
	DB          string            `json:"db"`
	Config      string            `json:"config"`
	Facts       string            `json:"facts"`
	Output      string            `json:"output"`
	Bin         string            `json:"bin"`
	DBOK        bool              `json:"db_ok"`
	LLMProvider string            `json:"llm_provider"`
	LLMCLI      string            `json:"llm_cli,omitempty"`
	PDFEngine   string            `json:"pdf_engine"`
	MCP         map[string]string `json:"mcp"`
	DesktopPATH bool              `json:"desktop_path_fix"`
}

// Doctor gathers install and environment diagnostics.
func Doctor(ctx context.Context, app *App) (DoctorReport, error) {
	p, err := paths.Resolve()
	if err != nil {
		return DoctorReport{}, err
	}
	if app != nil && app.Paths.Home != "" {
		p = app.Paths
	}
	rep := DoctorReport{
		Home:        p.Home,
		DB:          p.DB,
		Config:      p.Config,
		Facts:       p.Facts,
		Output:      p.Output,
		Bin:         p.Bin,
		PDFEngine:   doctorPDFEngine(p),
		MCP:         detectMCPRegistration(),
		DesktopPATH: os.Getenv("APPLYDIFF_DESKTOP") == "1",
	}
	if app != nil && app.Store != nil {
		rep.DBOK = app.Store.Ping(ctx) == nil
	}
	cfg := llm.ConfigFromEnv()
	rep.LLMProvider = cfg.Provider
	rep.LLMCLI = cfg.CLI
	return rep, nil
}

func detectMCPRegistration() map[string]string {
	out := map[string]string{
		"cursor": "not_registered",
		"claude": "not_registered",
		"codex":  "not_registered",
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return out
	}
	checks := []struct {
		key  string
		path string
	}{
		{"cursor", filepath.Join(home, ".cursor", "mcp.json")},
		{"claude", filepath.Join(home, ".claude", ".mcp.json")},
		{"codex", filepath.Join(home, ".codex", "mcp.json")},
	}
	for _, c := range checks {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}
		if mcpConfigReferencesApplydiff(data) {
			out[c.key] = "registered"
		}
	}
	return out
}

func mcpConfigReferencesApplydiff(data []byte) bool {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return strings.Contains(strings.ToLower(string(data)), "applydiff")
	}
	raw, _ := json.Marshal(v)
	return strings.Contains(strings.ToLower(string(raw)), "applydiff")
}

// InitData prepares the per-user folder and default config without opening the DB.
func InitData() (paths.Paths, error) {
	p, err := paths.Resolve()
	if err != nil {
		return paths.Paths{}, err
	}
	if err := paths.EnsureDefaults(p); err != nil {
		return paths.Paths{}, err
	}
	return p, nil
}

// OutputRoot returns the PDF output directory (env overrides paths).
func (a *App) OutputRoot() string {
	if v := os.Getenv("OUTPUT_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("EXPORT_ROOT"); v != "" {
		return v
	}
	if a != nil && a.Paths.Output != "" {
		return a.Paths.Output
	}
	return "output"
}

// DetectCLIOnPATH is used by doctor text output to show which agent binary was found.
func DetectCLIOnPATH() string {
	return llm.DetectCLI()
}

// PDFEngineProbe re-exports export detection for CLI text mode.
func PDFEngineProbe() string {
	return export.DetectPDFEngine()
}

func doctorPDFEngine(p paths.Paths) string {
	if eng, err := export.ResolveEngine(p); err == nil {
		return eng.Name()
	}
	return export.DetectPDFEngine()
}
