package fourdayweek

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestParsePage(t *testing.T) {
	f, err := os.Open("testdata/page1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	jobs, hasMore, err := parsePage(f)
	if err != nil {
		t.Fatal(err)
	}
	if hasMore {
		t.Fatal("expected no more pages")
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs: %d", len(jobs))
	}
	j := jobs[0]
	if j.Title != "Backend Engineer" || j.Company != "Acme" {
		t.Fatalf("unexpected job: %+v", j)
	}
	if j.WeeklyDays != 4 || j.Source != SourceID {
		t.Fatalf("weekly/source: %+v", j)
	}
	if j.URL != "https://4dayweek.io/jobs/test-job-1" {
		t.Fatalf("url: %s", j.URL)
	}
}

func TestFetchFiltersQuery(t *testing.T) {
	body, err := os.ReadFile("testdata/page1.json")
	if err != nil {
		t.Fatal(err)
	}
	src := &Source{
		Client: &roundTripper{body: body},
	}
	out, err := src.Fetch(context.Background(), []string{"kotlin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected filter miss, got %d", len(out))
	}
	out, err = src.Fetch(context.Background(), []string{"backend"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 job, got %d", len(out))
	}
}

type roundTripper struct {
	body []byte
}

func (r *roundTripper) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(r.body)),
		Header:     make(http.Header),
	}, nil
}
