package export

import (
	"strings"
	"testing"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestPDFOutputPathSlugified(t *testing.T) {
	at := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	j := job.Job{
		Company: "Acme Corp!",
		Title:   "Senior Go Engineer",
	}
	got := PDFOutputPath("output", j, at)
	want := "output/2026-10-17-acme-corp-senior-go-engineer.pdf"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if strings.Contains(got, "/") && strings.Count(got, "/") > 1 {
		t.Fatalf("expected flat output path, got %q", got)
	}
}
