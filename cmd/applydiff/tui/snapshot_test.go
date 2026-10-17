package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/features"
	"github.com/pahlevikun/applydiff/internal/interview"
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

func TestInterviewScreenSnapshot(t *testing.T) {
	if !features.InterviewPrepEnabled {
		t.Skip("interview prep disabled")
	}
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
		ID: "job-iv", Company: "Acme", Title: "Platform Engineer", URL: "https://example.com",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "discovered",
		Description: "distributed systems and caching",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	app := &runtime.App{Store: st, Config: &jobconfig.Config{}}
	m := initialModel(app)
	m.width = 110
	m.height = 28
	m.screen = screenInterview
	m.selected = j.ID
	fw := interview.FrameworkByID("system-design")
	if fw == nil {
		t.Fatal("framework missing")
	}
	m.ivFramework = fw
	m.ivLoop = []string{"recruiter screen", "technical screen", "coding"}
	m.ivTopics = []string{"caching", "scalability"}
	m.ensureIVSteps(fw)
	m.ivStepIdx = 1
	out := m.View()
	if !strings.Contains(out, "Interview") {
		t.Fatalf("expected Interview screen header, got:\n%s", out)
	}
	if !strings.Contains(out, "Clarify requirements") && !strings.Contains(out, "Estimate scale") {
		t.Fatalf("expected step detail, got:\n%s", out)
	}
	if !strings.Contains(out, "Loop") {
		t.Fatalf("expected kit loop column, got:\n%s", out)
	}
}
