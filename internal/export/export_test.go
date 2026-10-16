package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	cvexport "github.com/pahlevikun/applydiff/internal/cv/export"
)

func TestExportProducesTeXAndDrafts(t *testing.T) {
	doc := sampleDoc()
	dir := t.TempDir()
	res, err := Build(Options{
		OutRoot: dir,
		Job: job.Job{ID: "j1", Company: "Acme", Title: "Engineer"},
		Doc:     doc,
		Facts:   &cv.Facts{Facts: map[string]string{"uptime": "99.9%"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	texBytes, err := os.ReadFile(res.CVTeX)
	if err != nil {
		t.Fatal(err)
	}
	tex := string(texBytes)
	if !cvexport.ContainsSections(tex) {
		t.Fatalf("missing sections in cv tex")
	}
	cover, err := os.ReadFile(res.CoverLetter)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cover), "Acme") {
		t.Fatalf("cover missing company")
	}
	proof, err := os.ReadFile(res.ProofPage)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(proof), "Shipped APIs") {
		t.Fatalf("proof missing bullet text")
	}
}

func TestExportRejectsInventedNumber(t *testing.T) {
	doc := sampleDoc()
	doc.Summary.Text = "Improved uptime to 99.9% across services"
	dir := t.TempDir()
	_, err := Build(Options{
		OutRoot: dir,
		Job:     job.Job{ID: "j2", Company: "Acme", Title: "Engineer"},
		Doc:     doc,
		Facts:   &cv.Facts{Facts: map[string]string{"other": "10%"}},
	})
	if err == nil {
		t.Fatal("expected facts gate error for unapproved 99.9%")
	}
}

func TestCoverLetterUsesOnlyCVNumbers(t *testing.T) {
	doc := sampleDoc()
	doc.Summary.Text = "Backend engineer with 5 years in Go"
	facts := &cv.Facts{Facts: map[string]string{"yrs": "5 years"}}
	if err := facts.GateNumberIn(coverPlainText(doc, job.Job{Company: "Co", Title: "Role"})); err != nil {
		t.Fatalf("cover should not invent numbers: %v", err)
	}
	if strings.Contains(doc.Summary.Text, "999") {
		t.Fatal("invented number in cover source")
	}
}

func sampleDoc() *cv.Document {
	return &cv.Document{
		Person: "p",
		Contact: cv.Contact{
			Name: "Ada Lovelace", Headline: "Senior Engineer", Email: "ada@example.com",
			Site: "https://example.com", LinkedIn: "https://linkedin.com/in/ada", GitHub: "https://github.com/ada",
			Location: "Remote",
		},
		Summary: cv.Summary{ID: "summary", Text: "Backend engineer focused on reliable systems."},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Shipped APIs for payments"}},
		}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Led migration to Go"}},
	}
}

func TestExportSnapshotFrozen(t *testing.T) {
	doc := sampleDoc()
	dir := t.TempDir()
	res, err := Build(Options{
		OutRoot: dir,
		Job:     job.Job{ID: "j3", Company: "Acme", Title: "Eng"},
		Doc:     doc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CVSnapshot == "" {
		t.Fatal("expected cv snapshot json")
	}
	loaded, err := cv.LoadJSON([]byte(res.CVSnapshot))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Contact.Name != "Ada Lovelace" {
		t.Fatalf("snapshot name %q", loaded.Contact.Name)
	}
	_ = filepath.Base(res.CVTeX)
}
