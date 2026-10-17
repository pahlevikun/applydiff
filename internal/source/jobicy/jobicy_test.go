package jobicy

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
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Source != SourceID || jobs[0].Company == "" {
		t.Fatalf("unexpected job: %+v", jobs[0])
	}
}
