package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Paths holds per-user data locations shared by CLI, serve, TUI, and desktop.
type Paths struct {
	Home       string
	DB         string
	Config     string
	Facts      string
	Output     string
	Bin        string
	MasterJSON string
	MasterTeX  string
}

// Resolve picks data locations: explicit env vars, then legacy ./applydiff.db in cwd,
// otherwise the OS application data directory.
func Resolve() (Paths, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Paths{}, err
	}
	base, err := resolveHome(cwd)
	if err != nil {
		return Paths{}, err
	}
	masterJSON, masterTeX, cvFacts := resolveCVSources(cwd)
	factsPath := filepath.Join(base, "facts.json")
	if fileExists(cvFacts) {
		factsPath = cvFacts
	}
	p := Paths{
		Home:       base,
		DB:         filepath.Join(base, "applydiff.db"),
		Config:     filepath.Join(base, "jobs", "config.yaml"),
		Facts:      factsPath,
		Output:     filepath.Join(base, "output"),
		Bin:        filepath.Join(base, "bin"),
		MasterJSON: masterJSON,
		MasterTeX:  masterTeX,
	}
	applyEnvOverrides(&p)
	return p, nil
}

// Canonical repo-relative CV paths (override with MASTER_CV_JSON / MASTER_CV_TEX / FACTS_PATH).
const (
	DirCV          = "cv"
	FileMasterJSON = "master_cv.json"
	FileMasterTeX  = "master_cv.tex"
	FileCVFacts    = "facts.json"
)

func resolveCVSources(cwd string) (jsonPath, texPath, factsPath string) {
	cvDir := filepath.Join(cwd, DirCV)
	jsonPath = filepath.Join(cvDir, FileMasterJSON)
	texPath = filepath.Join(cvDir, FileMasterTeX)
	factsPath = filepath.Join(cvDir, FileCVFacts)
	if !fileExists(jsonPath) {
		jsonPath = filepath.Join(cwd, FileMasterJSON)
	}
	if !fileExists(texPath) {
		texPath = filepath.Join(cwd, FileMasterTeX)
	}
	if !fileExists(factsPath) {
		factsPath = filepath.Join(cwd, FileCVFacts)
	}
	return jsonPath, texPath, factsPath
}

func resolveHome(cwd string) (string, error) {
	if fileExists(filepath.Join(cwd, "applydiff.db")) {
		return cwd, nil
	}
	return appDataDir()
}

func applyEnvOverrides(p *Paths) {
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		p.DB = v
	}
	if v := os.Getenv("JOBS_CONFIG"); v != "" {
		p.Config = v
	}
	if v := os.Getenv("FACTS_PATH"); v != "" {
		p.Facts = v
	}
	if v := os.Getenv("OUTPUT_ROOT"); v != "" {
		p.Output = v
	}
	if v := os.Getenv("APPLYDIFF_BIN"); v != "" {
		p.Bin = v
	}
	if v := os.Getenv("APPLYDIFF_HOME"); v != "" {
		p.Home = v
	}
	if v := os.Getenv("MASTER_CV_JSON"); v != "" {
		p.MasterJSON = v
	}
	if v := os.Getenv("MASTER_CV_TEX"); v != "" {
		p.MasterTeX = v
	}
}

func appDataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "applydiff"), nil
	case "windows":
		dir := os.Getenv("APPDATA")
		if dir == "" {
			return "", fmt.Errorf("paths: APPDATA is not set")
		}
		return filepath.Join(dir, "applydiff"), nil
	default:
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return filepath.Join(dir, "applydiff"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "applydiff"), nil
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
