package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func TestRetailerKeepsDecisionsAndMarksStale(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "job-retailor"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "Co", Title: "Eng", URL: "https://example.com/y",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "matched",
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveMaster(ctx, &cv.Document{
		Person: "p", Contact: cv.Contact{Name: "A", Headline: "H", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "old"},
	}); err != nil {
		t.Fatal(err)
	}
	h1 := []tailor.Hunk{{
		FieldID: "summary", Section: "summary", Before: "old", After: "new", Rationale: "fit",
	}}
	prop, err := st.SaveProposal(ctx, jobID, h1)
	if err != nil {
		t.Fatal(err)
	}
	hash := hunkdecision.ContentHash(prop.Hunks[0])
	if _, err := st.UpsertHunkDecision(ctx, jobID, 0, hunkdecision.DecisionAccept, "", hash); err != nil {
		t.Fatal(err)
	}
	prop2, err := st.SaveProposal(ctx, jobID, []tailor.Hunk{{
		FieldID: "summary", Section: "summary", Before: "old", After: "newer", Rationale: "fit",
	}})
	if err != nil {
		t.Fatal(err)
	}
	stale, err := st.ApplyHunkDecisions(ctx, jobID, prop2.Hunks)
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 || !stale[0] {
		t.Fatalf("expected stale hunk, stale=%v", stale)
	}
	if prop2.Hunks[0].Status != tailor.StatusPending {
		t.Fatalf("stale hunk should be pending, got %s", prop2.Hunks[0].Status)
	}
	prop3, err := st.SaveProposal(ctx, jobID, h1)
	if err != nil {
		t.Fatal(err)
	}
	if prop3.Hunks[0].Status != tailor.StatusAccepted {
		t.Fatalf("unchanged hunk should stay accepted, got %s", prop3.Hunks[0].Status)
	}
}
