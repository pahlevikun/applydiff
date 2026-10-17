package remotive

import (
	"os"
	"strings"
	"testing"
)

func TestParseJobsFixture(t *testing.T) {
	f, err := os.Open("testdata/jobs_min.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	jobs, err := parseJobs(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Source != SourceID || j.Company != "Acme Remote" {
		t.Fatalf("unexpected job: %+v", j)
	}
	if j.Remote != "yes" || j.Arrangement != "remote" {
		t.Fatalf("expected remote job, got remote=%q arrangement=%q", j.Remote, j.Arrangement)
	}
	if !strings.Contains(j.Title, "Backend") {
		t.Fatal("expected backend title")
	}
}
