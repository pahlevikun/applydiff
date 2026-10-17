package remoteok

import (
	"os"
	"testing"
)

func TestParseJobsFixture(t *testing.T) {
	f, err := os.Open("testdata/jobs.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	jobs, err := parseJobs(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) < 1 {
		t.Fatalf("expected jobs, got %d", len(jobs))
	}
	if jobs[0].Source != SourceID {
		t.Fatalf("unexpected source %q", jobs[0].Source)
	}
	if jobs[0].Title == "" || jobs[0].Company == "" {
		t.Fatalf("missing title/company: %+v", jobs[0])
	}
}
