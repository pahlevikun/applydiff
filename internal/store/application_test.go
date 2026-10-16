package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/application"
	"github.com/pahlevikun/applydiff/internal/job"
)

func TestApplicationTransition(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "app-job-1"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "C", Title: "T", URL: "https://x", Source: "s",
		PostedAt: "2026-01-01", Status: "discovered",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateApplication(ctx, jobID); err != nil {
		t.Fatal(err)
	}
	app, err := st.TransitionApplication(ctx, jobID, application.StatusMatched, "")
	if err != nil {
		t.Fatal(err)
	}
	if app.Status != application.StatusMatched {
		t.Fatalf("status %s", app.Status)
	}
	_, err = st.TransitionApplication(ctx, jobID, application.StatusSubmitted, "")
	if err == nil {
		t.Fatal("illegal transition should fail")
	}
}
