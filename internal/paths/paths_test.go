package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

func samePath(a, b string) bool {
	norm := func(p string) string {
		aa, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		if r, err := filepath.EvalSymlinks(aa); err == nil {
			return r
		}
		return aa
	}
	return norm(a) == norm(b)
}

func TestResolveLegacyCwd(t *testing.T) {
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	db := filepath.Join(dir, "applydiff.db")
	if err := os.WriteFile(db, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	p, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !samePath(p.Home, dir) {
		t.Fatalf("home: got %q want %q", p.Home, dir)
	}
	if !samePath(p.DB, db) {
		t.Fatalf("db: got %q want %q", p.DB, db)
	}
	wantCfg := filepath.Join(p.Home, "jobs", "config.yaml")
	if !samePath(p.Config, wantCfg) {
		t.Fatalf("config: got %q want %q", p.Config, wantCfg)
	}
}

func TestResolveEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	t.Setenv("DATABASE_PATH", filepath.Join(dir, "custom.db"))
	t.Setenv("JOBS_CONFIG", filepath.Join(dir, "cfg.yaml"))

	p, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if p.DB != filepath.Join(dir, "custom.db") {
		t.Fatalf("db override: %q", p.DB)
	}
	if p.Config != filepath.Join(dir, "cfg.yaml") {
		t.Fatalf("config override: %q", p.Config)
	}
}

func TestResolveAppDataWhenNoLegacyDB(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	p, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	switch runtime.GOOS {
	case "darwin":
		if !filepath.IsAbs(p.Home) || filepath.Base(p.Home) != "applydiff" {
			t.Fatalf("unexpected home %q", p.Home)
		}
	case "windows":
		if filepath.Base(p.Home) != "applydiff" {
			t.Fatalf("unexpected home %q", p.Home)
		}
	default:
		if filepath.Base(p.Home) != "applydiff" {
			t.Fatalf("unexpected home %q", p.Home)
		}
	}
	if filepath.Base(p.DB) != "applydiff.db" {
		t.Fatalf("db %q", p.DB)
	}
}

func TestEnsureDefaultsWritesConfig(t *testing.T) {
	dir := t.TempDir()
	p := Paths{
		Home:   dir,
		DB:     filepath.Join(dir, "applydiff.db"),
		Config: filepath.Join(dir, "jobs", "config.yaml"),
		Facts:  filepath.Join(dir, "facts.json"),
		Output: filepath.Join(dir, "output"),
		Bin:    filepath.Join(dir, "bin"),
	}
	if err := EnsureDefaults(p); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p.Config)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || len(data) != len(DefaultJobsConfigYAML()) {
		t.Fatalf("config length %d", len(data))
	}
}
