package job

import (
	"strings"
	"testing"
)

func TestDedupeCollapsesSameURLCompanyRole(t *testing.T) {
	a := Job{
		Company: "Acme",
		Title:   "Senior Engineer",
		URL:     "https://boards.greenhouse.io/acme/jobs/1?utm=1",
	}
	b := Job{
		Company: "Acme",
		Title:   "Senior Engineer",
		URL:     "https://www.boards.greenhouse.io/acme/jobs/1/",
	}
	got := Dedupe([]Job{a, b})
	if len(got) != 1 {
		t.Fatalf("expected 1 job after dedupe, got %d", len(got))
	}
}

func TestIDFromDedupeKeyStableAndSafe(t *testing.T) {
	key := DedupeKey(Job{Company: "Acme", Title: "Eng", URL: "https://example.com/j"})
	id := IDFromDedupeKey(key)
	if id == "" || strings.Contains(id, "\x00") {
		t.Fatalf("bad id %q", id)
	}
	if IDFromDedupeKey(key) != id {
		t.Fatal("expected stable id")
	}
}

func TestNormalizeURLStripsQueryAndWWW(t *testing.T) {
	in := "https://WWW.Example.com/path/?utm=1"
	want := "https://example.com/path"
	if got := NormalizeURL(in); got != want {
		t.Fatalf("NormalizeURL: got %q want %q", got, want)
	}
}
