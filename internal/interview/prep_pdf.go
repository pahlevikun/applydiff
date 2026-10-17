package interview

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/export/fonts"
)

// WritePrepPDF renders markdown content to PDF via the shared export PDF engine.
func WritePrepPDF(mdContent string, pdfPath string) error {
	workDir, err := os.MkdirTemp("", "applydiff-interview-tex-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	texName := "interview-prep.tex"
	texPath := filepath.Join(workDir, texName)
	fontPath, err := fonts.DejaVuSansPath(workDir)
	if err != nil {
		return err
	}
	tex := markdownToLaTeX(mdContent, fontPath)
	if err := os.WriteFile(texPath, []byte(tex), 0o644); err != nil {
		return err
	}
	return export.CompilePDF(texPath, workDir, pdfPath)
}

func markdownToLaTeX(md string, dejaVuPath string) string {
	var b strings.Builder
	b.WriteString("\\documentclass[11pt]{article}\n")
	b.WriteString("\\usepackage[margin=1in]{geometry}\n")
	b.WriteString("\\usepackage{fontspec}\n")
	b.WriteString(fonts.MainfontSpec(dejaVuPath))
	b.WriteString("\\usepackage{hyperref}\n")
	b.WriteString("\\begin{document}\n")
	b.WriteString("\\begin{verbatim}\n")
	for _, line := range strings.Split(md, "\n") {
		b.WriteString(escapeLaTeXVerbatim(line))
		b.WriteString("\n")
	}
	b.WriteString("\\end{verbatim}\n")
	b.WriteString("\\end{document}\n")
	return b.String()
}

func escapeLaTeXVerbatim(line string) string {
	// verbatim handles most chars; still escape end of verbatim marker.
	return strings.ReplaceAll(line, "\\end{verbatim}", "\\end{verbatim}")
}
