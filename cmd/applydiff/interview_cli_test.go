package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestInterviewHelpListsSubcommands(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"interview", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{"kit", "mock", "drill", "story", "score", "list"} {
		if !containsWord(out, sub) {
			t.Fatalf("interview help missing %q:\n%s", sub, out)
		}
	}
}

func TestInterviewListHelp(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"interview", "list", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !containsWord(buf.String(), "job-key") {
		t.Fatalf("expected list help to mention job-key:\n%s", buf.String())
	}
}

func moduleRoot(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func TestInterviewKitMissingJobExits2(t *testing.T) {
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
	cmd := exec.Command(bin, "interview", "kit", "missing-job-key")
	cmd.Env = append(os.Environ(),
		"DATABASE_PATH="+dbPath,
		"JOBS_CONFIG="+filepath.Join(root, "jobs", "config.yaml"),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr == nil {
		t.Fatal("expected non-zero exit")
	}
	exitErr, ok := runErr.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError: %v", runErr)
	}
	if exitErr.ExitCode() != runtime.ExitNoMatch {
		t.Fatalf("exit code %d want %d stderr=%s", exitErr.ExitCode(), runtime.ExitNoMatch, stderr.String())
	}
}
