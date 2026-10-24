package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func TestHunkDecisionUpsertIdempotent(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "job-hunk-dec"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "Co", Title: "Eng", URL: "https://example.com/x",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "matched",
	}); err != nil {
		t.Fatal(err)
	}
	h := tailor.Hunk{FieldID: "summary", Section: "summary", Before: "a", After: "b", Rationale: "r"}
	hash := hunkdecision.ContentHash(h)
	rec1, err := st.UpsertHunkDecision(ctx, jobID, 0, hunkdecision.DecisionAccept, "", hash)
	if err != nil {
		t.Fatal(err)
	}
	rec2, err := st.UpsertHunkDecision(ctx, jobID, 0, hunkdecision.DecisionAccept, "", hash)
	if err != nil {
		t.Fatal(err)
	}
	if rec1.Decision != rec2.Decision || rec1.ContentHash != rec2.ContentHash {
		t.Fatalf("idempotent upsert mismatch: %+v vs %+v", rec1, rec2)
	}
	list, err := st.ListHunkDecisions(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(list))
	}
}
