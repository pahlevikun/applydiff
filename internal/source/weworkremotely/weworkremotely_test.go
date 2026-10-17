package weworkremotely

import (
	"os"
	"strings"
	"testing"
)

func TestParseFeedFixture(t *testing.T) {
	f, err := os.Open("testdata/feed.rss")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	jobs, err := parseFeed(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Company != "Acme Corp" || j.Source != SourceID {
		t.Fatalf("unexpected job: %+v", j)
	}
	if !strings.Contains(j.Title, "Backend") {
		t.Fatal("expected backend in title")
	}
}
