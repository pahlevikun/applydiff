package greenhouse

import (
	"os"
	"strings"
	"testing"
)

func TestParseJobsFixture(t *testing.T) {
	f, err := os.Open("testdata/stripe_jobs.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	jobs, err := parseJobs(f, Board{Token: "stripe", Company: "Stripe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Company != "Stripe" || jobs[0].Source != SourceID {
		t.Fatalf("unexpected first job: %+v", jobs[0])
	}
	if jobs[0].Remote != "yes" {
		t.Fatalf("expected remote yes, got %q", jobs[0].Remote)
	}
	if !strings.Contains(jobs[0].Title, "Senior") {
		t.Fatalf("expected senior title in fixture")
	}
}

func TestMatchesQueries(t *testing.T) {
	if !matchesQueries("Senior Backend Engineer", []string{"backend"}) {
		t.Fatal("expected match")
	}
	if matchesQueries("Marketing Coordinator", []string{"backend"}) {
		t.Fatal("expected no match")
	}
}
