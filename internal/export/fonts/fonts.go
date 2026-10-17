package fonts

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed DejaVuSans.ttf
var dejaVuSans []byte

// DejaVuSansPath writes the bundled font next to workDir and returns its path.
func DejaVuSansPath(workDir string) (string, error) {
	dir := filepath.Join(workDir, ".applydiff-fonts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "DejaVuSans.ttf")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.WriteFile(path, dejaVuSans, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// MainfontSpec returns a fontspec setmainfont line for the given TTF path.
func MainfontSpec(ttfPath string) string {
	dir := filepath.ToSlash(filepath.Dir(ttfPath))
	return "\\setmainfont{DejaVu Sans}[Path=" + dir + "/,Extension=.ttf,UprightFont=DejaVuSans]\n"
}
