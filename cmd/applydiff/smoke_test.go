package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/match"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestSmokePipeline(t *testing.T) {
	ctx := context.Background()
	root := moduleRoot(t)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "smoke.db")
	outRoot := filepath.Join(dir, "output")
	t.Setenv("OUTPUT_ROOT", outRoot)

	cfg := &jobconfig.Config{
		Filters: jobconfig.Filters{Spec: filter.Spec{
			MaxPostedDays: &filter.IntRule{Hard: false, Max: 365},
		}},
	}
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     dbPath,
		ConfigPath: filepath.Join(root, "testdata", "smoke", "jobs.yaml"),
		FactsPath:  filepath.Join(root, "examples", "cv", "facts.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	app.Config = cfg
	app.LLM = llm.FakeRunner{Response: smokeTailorJSON(root)}

	masterPath := filepath.Join(root, "examples", "cv", "master.json")
	if err := app.ImportMasterCV(ctx, masterPath, false); err != nil {
		t.Fatal(err)
	}
	st := app.Store

	j := job.Job{
		ID:          "smoke-job-1",
		Company:     "Acme Remote",
		Title:       "Senior Backend Engineer",
		URL:         "https://remotive.com/remote-jobs/software-development/senior-backend-engineer-1001",
		Source:      "remotive",
		PostedAt:    "2026-10-10T10:00:00Z",
		Remote:      "yes",
		Region:      "global",
		Status:      "discovered",
		Description: "Go backend role with distributed systems and Kafka.",
		Arrangement: "remote",
		Keywords:    []string{"golang", "api"},
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	j.DedupeKey = job.DedupeKey(j)

	views, err := app.ListJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) < 1 {
		t.Fatal("expected jobs list")
	}
	for _, v := range views {
		dk := v.Job.DedupeKey
		if dk == "" {
			dk = job.DedupeKey(v.Job)
		}
		if strings.TrimSpace(dk) == "" {
			t.Fatalf("missing dedupe_key on job %s", v.Job.ID)
		}
	}

	doc, err := app.MasterCV(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mres := match.Evaluate(match.Input{Job: j, CV: doc})
	if mres.Total < inbox.MatchMinScore {
		t.Fatalf("match score %d below inbox threshold", mres.Total)
	}
	if err := st.SaveMatchResult(ctx, mres); err != nil {
		t.Fatal(err)
	}

	inboxRes, err := inbox.List(ctx, st, inbox.ListOptions{State: "needs-decision"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range inboxRes.Data {
		if row.JobKey == j.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("job %s not in needs-decision inbox", j.ID)
	}

	prop, err := app.Tailor(ctx, j.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(prop.Hunks) < 1 {
		t.Fatalf("expected hunks, got %d", len(prop.Hunks))
	}

	prop, err = app.ApproveHunks(ctx, j.ID, nil, true, false, runtime.ActorHuman)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.PendingHunks(prop.Hunks) != 0 {
		t.Fatalf("expected all hunks approved, pending=%d", runtime.PendingHunks(prop.Hunks))
	}

	res, err := app.BuildPDF(ctx, j.ID, false, runtime.ActorHuman)
	if err != nil {
		if runtime.ExitCode(err) != runtime.ExitLaTeX {
			t.Fatalf("build: %v", err)
		}
		var latexErr runtime.ErrLaTeX
		if e, ok := err.(runtime.ErrLaTeX); ok {
			latexErr = e
		}
		if !strings.Contains(strings.ToLower(latexErr.Error()), "xelatex") &&
			!strings.Contains(strings.ToLower(latexErr.Error()), "tectonic") &&
			!strings.Contains(strings.ToLower(latexErr.Error()), "latex") {
			t.Fatalf("expected latex-related error, got %v", err)
		}
		return
	}
	if res.CVPDF == "" {
		t.Fatal("expected cv_pdf path")
	}
	if _, err := os.Stat(res.CVPDF); err != nil {
		t.Fatalf("pdf missing: %v", err)
	}
}

func smokeTailorJSON(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "testdata", "smoke", "tailor-hunks.json"))
	if err != nil {
		return `{"hunks":[]}`
	}
	return string(data)
}
