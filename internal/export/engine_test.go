package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/export/fonts"
	"github.com/pahlevikun/applydiff/internal/paths"
)

func TestResolveEngineEnvOverride(t *testing.T) {
	t.Setenv("APPLYDIFF_PDF_ENGINE", "tectonic")
	if _, err := exec.LookPath("tectonic"); err != nil {
		t.Skip("tectonic not on PATH")
	}
	eng, err := ResolveEngine(paths.Paths{})
	if err != nil {
		t.Fatal(err)
	}
	if eng.Name() != "tectonic" {
		t.Fatalf("got %q", eng.Name())
	}
}

func TestResolveEngineBinTectonic(t *testing.T) {
	if _, err := exec.LookPath("xelatex"); err == nil {
		t.Skip("xelatex on PATH would win over bin tectonic")
	}
	dir := t.TempDir()
	fake := filepath.Join(dir, "tectonic")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho tectonic\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	eng, err := ResolveEngine(paths.Paths{Bin: dir})
	if err != nil {
		t.Fatal(err)
	}
	if eng.Name() != "tectonic" {
		t.Fatalf("got %q", eng.Name())
	}
}

func TestDejaVuBundledFontExists(t *testing.T) {
	dir := t.TempDir()
	p, err := fonts.DejaVuSansPath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("missing bundled font: %v", p)
	}
	line := fonts.MainfontSpec(p)
	if !strings.Contains(line, "Path=") || !strings.Contains(line, "DejaVuSans") {
		t.Fatalf("unexpected fontspec line: %q", line)
	}
}
