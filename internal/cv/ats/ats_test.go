package ats

import (
	"testing"
)

func TestAnalyzeTeX(t *testing.T) {
	tex := "\\section*{Summary}\\section*{Technical Skills}\\section*{Work Experience} Go ABAC 40\\%"
	r := AnalyzeTeX(tex, []string{"Go", "ABAC", "Rust"})
	if !r.OrderOK {
		t.Fatal("expected reading order ok")
	}
	if len(r.KeywordsMissing) != 1 || r.KeywordsMissing[0] != "Rust" {
		t.Fatalf("unexpected missing: %v", r.KeywordsMissing)
	}
}
