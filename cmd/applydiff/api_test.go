package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/store"
)

func testServer(st *store.Store, cfg *jobconfig.Config, runner *scan.Runner, llmRunner llm.Runner) *server {
	if cfg == nil {
		cfg = &jobconfig.Config{}
	}
	app := &runtime.App{
		Store:  st,
		Config: cfg,
		Runner: runner,
		LLM:    llmRunner,
	}
	if app.Runner == nil {
		app.Runner = &scan.Runner{Store: st, Config: cfg}
	}
	if app.LLM == nil {
		app.LLM = llm.FakeRunner{Response: `{"hunks":[]}`}
	}
	return &server{app: app}
}

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
		Filters: jobconfig.Filters{Spec: filter.Spec{
			Regions:       &filter.CategoryRule{Hard: true, Values: []string{"global"}},
			MaxPostedDays: &filter.IntRule{Hard: true, Max: 30},
		}},
	}
	srv := testServer(st, cfg, &scan.Runner{
		Store:  st,
		Config: cfg,
		Client: &alwaysAliveClient{},
	}, nil)

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
	var jobs []job.JobView
	if err := json.NewDecoder(listRec.Body).Decode(&jobs); err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if !jobs[0].Filter.Keep {
		t.Fatalf("expected keep true, reasons %v", jobs[0].Filter.Reasons)
	}
}

func TestAPIFilterPresetsRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	srv := testServer(st, &jobconfig.Config{}, nil, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/filters/presets", srv.handleFilterPresets)
	mux.HandleFunc("/api/filters/presets/", srv.handleFilterPresets)

	body := `{"name":"Test","spec":{"arrangement":{"hard":true,"values":["remote"]}}}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/filters/presets", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("post %d %s", postRec.Code, postRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.NewDecoder(postRec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	applyReq := httptest.NewRequest(http.MethodPost, "/api/filters/presets/"+id+"/apply", nil)
	applyRec := httptest.NewRecorder()
	mux.ServeHTTP(applyRec, applyReq)
	if applyRec.Code != http.StatusOK {
		t.Fatalf("apply %d", applyRec.Code)
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

	srv := testServer(st, &jobconfig.Config{}, nil, nil)
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

func TestAPITailorRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	jobID := "job-tailor-1"
	j := job.Job{
		ID:          jobID,
		Company:     "Acme",
		Title:       "Go Engineer",
		URL:         "https://example.com/1",
		Source:      "stub",
		PostedAt:    "2026-10-10T10:00:00Z",
		Remote:      "yes",
		Region:      "global",
		Status:      "matched",
		Description: "Looking for Go and distributed systems experience.",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	doc := &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Eng", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "Go engineer"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
	if err := st.SaveMaster(ctx, doc); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveFacts(ctx, &cv.Facts{Facts: map[string]string{"m": "40%"}}); err != nil {
		t.Fatal(err)
	}

	canned := `{"hunks":[{"field_id":"exp.1.b1","section":"experience","before":"Built APIs","after":"Architected distributed Go APIs","rationale":"JD fit","evidence":[]}]}`
	srv := testServer(st, &jobconfig.Config{}, nil, llm.FakeRunner{Response: canned})
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tailor/", srv.handleTailor)
	mux.HandleFunc("/api/hunks/", srv.handleHunkAction)

	postReq := httptest.NewRequest(http.MethodPost, "/api/tailor/"+jobID, nil)
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("tailor post %d %s", postRec.Code, postRec.Body.String())
	}
	var prop map[string]interface{}
	if err := json.NewDecoder(postRec.Body).Decode(&prop); err != nil {
		t.Fatal(err)
	}
	hunks, ok := prop["hunks"].([]interface{})
	if !ok || len(hunks) != 1 {
		t.Fatalf("expected 1 hunk in proposal: %#v", prop["hunks"])
	}
	h0 := hunks[0].(map[string]interface{})
	hunkID := h0["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/api/tailor/"+jobID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("tailor get %d", getRec.Code)
	}

	accReq := httptest.NewRequest(http.MethodPost, "/api/hunks/"+hunkID+"/accept", nil)
	accRec := httptest.NewRecorder()
	mux.ServeHTTP(accRec, accReq)
	if accRec.Code != http.StatusOK {
		t.Fatalf("accept %d %s", accRec.Code, accRec.Body.String())
	}

	branch, err := st.LoadBranch(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	text, _ := branch.Field("exp.1.b1")
	if text != "Architected distributed Go APIs" {
		t.Fatalf("branch cv: %q", text)
	}
}

func TestAPIApplicationsAndExport(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	jobID := "job-export-1"
	j := job.Job{
		ID: jobID, Company: "Acme", Title: "Go Engineer", URL: "https://example.com/1",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "global", Status: "matched",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	doc := &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Eng", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "Go engineer"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
	if err := st.SaveMaster(ctx, doc); err != nil {
		t.Fatal(err)
	}

	srv := testServer(st, &jobconfig.Config{Queries: []string{"golang remote"}}, nil, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/applications", srv.handleApplicationsList)
	mux.HandleFunc("/api/applications/", srv.handleApplicationStatus)
	mux.HandleFunc("/api/export/", srv.handleExport)
	mux.HandleFunc("/api/config", srv.handleConfig)

	transReq := httptest.NewRequest(http.MethodPost, "/api/applications/"+jobID+"/status", strings.NewReader(`{"status":"matched"}`))
	transReq.Header.Set("Content-Type", "application/json")
	transRec := httptest.NewRecorder()
	mux.ServeHTTP(transRec, transReq)
	if transRec.Code != http.StatusOK {
		t.Fatalf("transition %d %s", transRec.Code, transRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/applications", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list %d", listRec.Code)
	}

	exportReq := httptest.NewRequest(http.MethodPost, "/api/export/"+jobID, nil)
	exportRec := httptest.NewRecorder()
	mux.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("export %d %s", exportRec.Code, exportRec.Body.String())
	}
	app, err := st.GetApplication(ctx, jobID)
	if err != nil || app.CVSnapshot == "" {
		t.Fatalf("expected frozen cv snapshot on application")
	}
	if err := st.SaveMaster(ctx, &cv.Document{
		Person: "p", Contact: cv.Contact{Name: "Changed"},
		Summary: cv.Summary{ID: "summary", Text: "x"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go"}}},
		Experience: []cv.Experience{{ID: "exp.1", Company: "C", Title: "T", Start: "2020", Bullets: []cv.Bullet{{ID: "b", Text: "t"}}}},
		Education: []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
	}); err != nil {
		t.Fatal(err)
	}
	app2, _ := st.GetApplication(ctx, jobID)
	if app2.CVSnapshot != app.CVSnapshot {
		t.Fatal("snapshot should stay frozen when master changes")
	}

	cfgReq := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	cfgRec := httptest.NewRecorder()
	mux.ServeHTTP(cfgRec, cfgReq)
	if cfgRec.Code != http.StatusOK {
		t.Fatalf("config %d", cfgRec.Code)
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
