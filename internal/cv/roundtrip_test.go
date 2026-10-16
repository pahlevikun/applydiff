package cv_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/cv/export"
	"github.com/pahlevikun/applydiff/internal/cv/importer"
)

func TestTexRoundTripPreservesMetrics(t *testing.T) {
	texPath := filepath.Join("..", "..", "fixtures", "resume_ats_v2.tex")
	doc, facts, err := importer.ImportTex(texPath)
	if err != nil {
		t.Fatal(err)
	}
	if facts.Facts["abac_impl"] != "40%" {
		t.Fatalf("expected abac_impl fact, got %#v", facts.Facts)
	}

	out := export.ToTeX(doc)
	if !export.ContainsSections(out) {
		t.Fatal("exported tex missing required sections")
	}
	contentNums := cv.ExtractNumbers(cvPlainText(doc))
	outNums := cv.ExtractNumbers(out)
	for _, n := range contentNums {
		if !numberPresent(n, outNums) {
			t.Fatalf("metric %q missing from export; content=%v out=%v", n, contentNums, outNums)
		}
	}
}

func numberPresent(n string, outNums []string) bool {
	norm := strings.TrimSuffix(n, "%")
	for _, o := range outNums {
		if o == n {
			return true
		}
		if strings.TrimSuffix(o, "%") == norm {
			return true
		}
	}
	return false
}

func cvPlainText(doc *cv.Document) string {
	var b strings.Builder
	b.WriteString(doc.Summary.Text)
	for _, sg := range doc.Skills {
		b.WriteString(strings.Join(sg.Items, " "))
	}
	for _, exp := range doc.Experience {
		for _, bullet := range exp.Bullets {
			b.WriteString(bullet.Text)
		}
	}
	for _, ach := range doc.Achievements {
		b.WriteString(ach.Text)
	}
	b.WriteString(doc.Contact.Location)
	b.WriteString(doc.Contact.TimezoneLine)
	return b.String()
}
