package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cvexport "github.com/pahlevikun/applydiff/internal/cv/export"
	"github.com/pahlevikun/applydiff/internal/job"
)

// PDFOutputPath is output/<YYYY-MM-DD>-<company>-<role>.pdf (slugified company and role).
func PDFOutputPath(outRoot string, j job.Job, at time.Time) string {
	if outRoot == "" {
		outRoot = "output"
	}
	date := at.UTC().Format("2006-01-02")
	name := fmt.Sprintf("%s-%s-%s.pdf", date, slugify(j.Company), slugify(j.Title))
	return filepath.Join(outRoot, name)
}

// BuildPDF writes a single CV PDF for the job branch CV to PDFOutputPath.
func BuildPDF(opts Options) (pdfPath string, snapshot string, err error) {
	if opts.Doc == nil {
		return "", "", fmt.Errorf("export: cv document required")
	}
	if opts.Job.ID == "" {
		return "", "", fmt.Errorf("export: job id required")
	}
	root := opts.OutRoot
	if root == "" {
		root = "output"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", "", err
	}
	at := time.Now().UTC()
	pdfPath = PDFOutputPath(root, opts.Job, at)

	workDir, err := os.MkdirTemp("", "applydiff-tex-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(workDir)

	texName := fmt.Sprintf("cv-%s.tex", slugify(opts.Job.Company))
	texPath := filepath.Join(workDir, texName)
	cvTex := cvexport.ToTeX(opts.Doc)
	if err := os.WriteFile(texPath, []byte(cvTex), 0o644); err != nil {
		return "", "", err
	}
	if err := gateText(opts.Facts, proofPlainText(opts.Doc)); err != nil {
		return "", "", fmt.Errorf("export: cv: %w", err)
	}
	if err := CompilePDF(texPath, workDir, pdfPath); err != nil {
		return "", "", err
	}
	snap, _ := json.Marshal(opts.Doc)
	return pdfPath, string(snap), nil
}
