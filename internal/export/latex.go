package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runXeLaTeXLocal(texPath, workDir, pdfPath string) error {
	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o755); err != nil {
		return err
	}
	base := strings.TrimSuffix(filepath.Base(texPath), filepath.Ext(texPath))
	outDir := filepath.Dir(pdfPath)
	for i := 0; i < 2; i++ {
		cmd := exec.Command("xelatex", "-interaction=nonstopmode", "-output-directory="+outDir, texPath)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("xelatex: %w: %s", err, truncate(string(out), 400))
		}
	}
	got := filepath.Join(outDir, base+".pdf")
	if got != pdfPath {
		if err := os.Rename(got, pdfPath); err != nil {
			return err
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
