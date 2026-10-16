package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/store"
)

type stubSource struct {
	jobs []job.Job
}

func (s stubSource) ID() string { return "stub" }

func (s stubSource) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	return s.jobs, nil
}

func TestAPIJobsAndScan(t *testing.T) {
	source.Register(stubSource{jobs: []job.Job{
		{
			Company:  "Acme",
			Title:    "Senior Engineer",
			URL:      "https://example.com/job/1",
			Source:   "stub",
			PostedAt: "2026-10-10T10:00:00Z",
			Remote:   "yes",
			Region:   "global",
			Status:   "discovered",
		},
	}})

	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	cfg := &jobconfig.Config{
		Queries: []string{},
		Filters: jobconfig.Filters{Regions: []string{"global"}, MaxPostedDays: 30},
	}
	srv := &server{
		store: st,
		cfg:   cfg,
		runner: &scan.Runner{
			Store:  st,
			Config: cfg,
			Client: &alwaysAliveClient{},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/jobs", srv.handleJobs)
	mux.HandleFunc("/api/scan", srv.handleScan)

	scanReq := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
	scanRec := httptest.NewRecorder()
	mux.ServeHTTP(scanRec, scanReq)
	if scanRec.Code != http.StatusOK {
		t.Fatalf("scan status %d body %s", scanRec.Code, scanRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/jobs", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("jobs status %d", listRec.Code)
	}
	var jobs []job.Job
	if err := json.NewDecoder(listRec.Body).Decode(&jobs); err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

type alwaysAliveClient struct{}

func (alwaysAliveClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Request:    req,
	}, nil
}
