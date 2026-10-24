package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func TestAPIJobHunksPersistence(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "job-hunks-api"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "Acme", Title: "Eng", URL: "https://example.com/h",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "global", Status: "matched",
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveMaster(ctx, &cv.Document{
		Person: "p", Contact: cv.Contact{Name: "A", Headline: "H", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "x"},
	}); err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveProposal(ctx, jobID, []tailor.Hunk{{
		FieldID: "summary", Section: "summary", Before: "x", After: "y", Rationale: "fit",
	}})
	if err != nil {
		t.Fatal(err)
	}
	srv := testServer(st, &jobconfig.Config{}, nil, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/jobs/", srv.handleJobByID)

	postReq := httptest.NewRequest(http.MethodPost, "/api/jobs/"+jobID+"/hunks/1", strings.NewReader(`{"decision":"accept"}`))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("post decision %d %s", postRec.Code, postRec.Body.String())
	}
	getReq := httptest.NewRequest(http.MethodGet, "/api/jobs/"+jobID+"/hunks", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get hunks %d", getRec.Code)
	}
	var sum map[string]interface{}
	if err := json.NewDecoder(getRec.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	if sum["decided"] != float64(1) {
		t.Fatalf("expected 1 decided, got %v", sum["decided"])
	}
	rec, err := st.GetHunkDecision(ctx, jobID, 0)
	if err != nil || rec.Decision != hunkdecision.DecisionAccept {
		t.Fatalf("persisted decision %+v err=%v", rec, err)
	}
}
