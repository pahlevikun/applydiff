package job

import "testing"

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

func TestNormalizeURLStripsQueryAndWWW(t *testing.T) {
	in := "https://WWW.Example.com/path/?utm=1"
	want := "https://example.com/path"
	if got := NormalizeURL(in); got != want {
		t.Fatalf("NormalizeURL: got %q want %q", got, want)
	}
}
