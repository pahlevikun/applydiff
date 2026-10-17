package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/store"
)

func TestFeedScreenSnapshot(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	j := job.Job{
		ID: "job-1", Company: "Acme", Title: "Platform Engineer", URL: "https://example.com",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "discovered",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	app := &runtime.App{Store: st, Config: &jobconfig.Config{}}
	m := initialModel(app)
	m.width = 100
	m.height = 24
	var jv job.JobView
	jv.Job = j
	jv.Filter.Keep = true
	jv.Filter.Score = 42
	m.jobs = []job.JobView{jv}
	out := m.View()
	if !strings.Contains(out, "Feed") {
		t.Fatalf("expected Feed screen header, got:\n%s", out)
	}
	if !strings.Contains(out, "Platform Engineer") {
		t.Fatalf("expected job title in feed, got:\n%s", out)
	}
	if !strings.Contains(out, "Presets") {
		t.Fatalf("expected presets rail, got:\n%s", out)
	}
}
