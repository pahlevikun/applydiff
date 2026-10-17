package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompilePDF runs xelatex (or docker fallback) to produce pdfPath from texPath.
func CompilePDF(texPath, workDir, pdfPath string) error {
	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o755); err != nil {
		return err
	}
	if _, err := exec.LookPath("xelatex"); err == nil {
		return runXeLaTeXLocal(texPath, workDir, pdfPath)
	}
	if dockerAvailable() {
		return runXeLaTeXDocker(texPath, workDir, pdfPath)
	}
	return fmt.Errorf("xelatex not found on PATH and Docker is unavailable; install TeX Live (xetex) or Docker, or use the applydiff container image")
}

func runXeLaTeXLocal(texPath, workDir, pdfPath string) error {
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

func runXeLaTeXDocker(texPath, workDir, pdfPath string) error {
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}
	base := filepath.Base(texPath)
	outDir := filepath.Dir(pdfPath)
	args := []string{
		"run", "--rm",
		"-v", workDir + ":/work",
		"-v", outDir + ":/out",
		"texlive/texlive",
		"xelatex", "-interaction=nonstopmode", "-output-directory=/out", "/work/" + base,
	}
	for i := 0; i < 2; i++ {
		cmd := exec.Command("docker", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("docker xelatex: %w: %s", err, truncate(string(out), 400))
		}
	}
	baseName := strings.TrimSuffix(base, filepath.Ext(base)) + ".pdf"
	got := filepath.Join(outDir, baseName)
	if got != pdfPath {
		if err := os.Rename(got, pdfPath); err != nil && !fileExists(pdfPath) {
			return err
		}
	}
	return nil
}

func dockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
