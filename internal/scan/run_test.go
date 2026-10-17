package scan_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/store"
)

type fakeSource struct {
	id   string
	jobs []job.Job
}

func (f fakeSource) ID() string { return f.id }

func (f fakeSource) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	return f.jobs, nil
}

type aliveClient struct{}

func (aliveClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
}

func TestScanDedupeRecordsMultipleOrigins(t *testing.T) {
	source.Reset()
	url := "https://example.com/jobs/senior-backend-go"
	source.Register(fakeSource{
		id: "alpha",
		jobs: []job.Job{{
			Company: "Acme", Title: "Senior Backend Engineer", URL: url,
			Source: "alpha", PostedAt: "2026-10-10", Remote: "yes", Region: "global", Status: "discovered",
		}},
	})
	source.Register(fakeSource{
		id: "beta",
		jobs: []job.Job{{
			Company: "Acme", Title: "Senior Backend Engineer", URL: url,
			Source: "beta", PostedAt: "2026-10-11", Remote: "yes", Region: "global", Status: "discovered",
		}},
	})

	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	cfg := &jobconfig.Config{
		Filters: jobconfig.Filters{Spec: filter.Spec{
			Regions:       &filter.CategoryRule{Hard: false, Values: []string{"global"}},
			MaxPostedDays: &filter.IntRule{Hard: false, Max: 365},
		}},
	}
	runner := &scan.Runner{Store: st, Config: cfg, Client: aliveClient{}}
	res, err := runner.Run(context.Background(), scan.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Stored != 1 {
		t.Fatalf("expected 1 stored job, got stored=%d deduped=%d", res.Stored, res.Deduped)
	}
	jobs, err := st.ListJobs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job in db, got %d", len(jobs))
	}
	if len(jobs[0].Origins) < 2 {
		t.Fatalf("expected both origins recorded, got %#v", jobs[0].Origins)
	}
}
