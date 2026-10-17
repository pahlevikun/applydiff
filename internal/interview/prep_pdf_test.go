package interview

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/job"
)

func TestMarkdownToLaTeXUsesBundledFont(t *testing.T) {
	tex := markdownToLaTeX("# Title\n\nBody", "/tmp/work/.applydiff-fonts/DejaVuSans.ttf")
	if !strings.Contains(tex, "fontspec") {
		t.Fatal("expected fontspec")
	}
	if !strings.Contains(tex, "Path=") || !strings.Contains(tex, "DejaVuSans") {
		t.Fatalf("expected bundled DejaVu path in tex: %s", tex)
	}
}

func TestTectonicCompileTemplates(t *testing.T) {
	if _, err := exec.LookPath("tectonic"); err != nil {
		t.Skip("tectonic not installed")
	}
	t.Setenv("APPLYDIFF_PDF_ENGINE", "tectonic")
	doc := &cv.Document{
		Contact: cv.Contact{Name: "Ada Lovelace", Headline: "Engineer", Email: "ada@example.com", Location: "Remote"},
		Summary: cv.Summary{Text: "Backend engineer."},
		Experience: []cv.Experience{{
			Company: "Co", Title: "Eng", Start: "2020",
			Bullets: []cv.Bullet{{Text: "Shipped APIs"}},
		}},
		Achievements: []cv.Achievement{{Text: "Led migration"}},
	}
	dir := t.TempDir()
	res, err := export.Build(export.Options{
		OutRoot: dir,
		Job:     job.Job{ID: "j1", Company: "Acme", Title: "Engineer"},
		Doc:     doc,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tex := range []string{res.CVTeX, res.CoverLetter, res.ProofPage} {
		pdf := filepath.Join(dir, strings.TrimSuffix(filepath.Base(tex), ".tex")+"-tectonic.pdf")
		if err := export.CompilePDF(tex, filepath.Dir(tex), pdf); err != nil {
			t.Fatalf("compile %s: %v", filepath.Base(tex), err)
		}
	}
	prepPDF := filepath.Join(dir, "prep.pdf")
	if err := WritePrepPDF("# Interview prep\n\nLine one.", prepPDF); err != nil {
		t.Fatalf("prep pdf: %v", err)
	}
	if _, err := os.Stat(prepPDF); err != nil {
		t.Fatal(err)
	}
}
