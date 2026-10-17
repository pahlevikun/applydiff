package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestInterviewExportHelpDocumentsPath(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"interview", "export", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "output/<YYYY-MM-DD>") || !strings.Contains(out, "interview.md") {
		t.Fatalf("export help should document output path:\n%s", out)
	}
}

func TestInterviewExportMissingJobExits2(t *testing.T) {
	root := moduleRoot(t)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "applydiff.db")
	ctx := context.Background()
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     dbPath,
		ConfigPath: filepath.Join(root, "jobs", "config.yaml"),
	})
	if err != nil {
		t.Fatal(err)
	}
	app.Close()

	bin := filepath.Join(dir, "applydiff")
	build := exec.Command("go", "build", "-o", bin, "./cmd/applydiff")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cmd := exec.Command(bin, "interview", "export", "missing-job-key")
	cmd.Env = append(os.Environ(),
		"DATABASE_PATH="+dbPath,
		"JOBS_CONFIG="+filepath.Join(root, "jobs", "config.yaml"),
		"OUTPUT_ROOT="+filepath.Join(dir, "output"),
	)
	runErr := cmd.Run()
	if runErr == nil {
		t.Fatal("expected non-zero exit")
	}
	exitErr, ok := runErr.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError: %v", runErr)
	}
	if exitErr.ExitCode() != runtime.ExitNoMatch {
		t.Fatalf("exit code %d want %d", exitErr.ExitCode(), runtime.ExitNoMatch)
	}
}

func TestInterviewExportWritesMarkdown(t *testing.T) {
	ctx := context.Background()
	root := moduleRoot(t)
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "applydiff.db")
	outRoot := filepath.Join(dir, "output")
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     dbPath,
		ConfigPath: filepath.Join(root, "jobs", "config.yaml"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	j := job.Job{
		ID: "job-export-iv", Company: "Acme", Title: "Platform Engineer", URL: "https://example.com",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "discovered",
		Description: "Build APIs with caching and queues.",
	}
	if err := app.Store.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	if _, err := app.GenerateInterviewKit(ctx, j.ID); err != nil {
		t.Fatal(err)
	}
	os.Setenv("OUTPUT_ROOT", outRoot)
	res, err := app.ExportInterviewPrep(ctx, j.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(res.Markdown, "-interview.md") {
		t.Fatalf("unexpected markdown path %q", res.Markdown)
	}
	data, err := os.ReadFile(res.Markdown)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "Interview prep") || !strings.Contains(body, "Story bank") {
		t.Fatalf("unexpected export body:\n%s", body)
	}
}
