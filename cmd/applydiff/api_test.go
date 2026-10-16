package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
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

func TestAPIMatchRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	jobID := "job-api-match-1"
	j := job.Job{
		ID:          jobID,
		Company:     "Acme",
		Title:       "Senior Go Engineer",
		URL:         "https://boards.greenhouse.io/acme/jobs/1",
		Source:      "stub",
		PostedAt:    "2026-10-10T10:00:00Z",
		Remote:      "yes",
		Region:      "global",
		Status:      "discovered",
		Description: "- Required: 5+ years experience with Go\n- Kubernetes required",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}

	doc := &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Senior Engineer", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "Go backend"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go", "Kubernetes"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Senior Engineer", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "APIs"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
	if err := st.SaveMaster(ctx, doc); err != nil {
		t.Fatal(err)
	}

	srv := &server{store: st, cfg: &jobconfig.Config{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/match/", srv.handleMatch)

	postReq := httptest.NewRequest(http.MethodPost, "/api/match/"+jobID, nil)
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("post match status %d body %s", postRec.Code, postRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/match/"+jobID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get match status %d", getRec.Code)
	}

	var postRes, getRes map[string]interface{}
	if err := json.NewDecoder(postRec.Body).Decode(&postRes); err != nil {
		t.Fatal(err)
	}
	if err := json.NewDecoder(getRec.Body).Decode(&getRes); err != nil {
		t.Fatal(err)
	}
	if postRes["total"] != getRes["total"] {
		t.Fatalf("total mismatch post=%v get=%v", postRes["total"], getRes["total"])
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
