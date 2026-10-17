package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestExportImportRoundTrip(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	srcDB := filepath.Join(dir, "source.db")
	st, err := New(srcDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := st.InsertJob(ctx, job.Job{
		Company:  "Round",
		Title:    "Trip",
		URL:      "https://example.com/r",
		Source:   "test",
		PostedAt: "2026-01-01",
		Status:   "discovered",
	}); err != nil {
		t.Fatal(err)
	}
	exportPath := filepath.Join(dir, "export.db")
	if err := st.ExportDB(ctx, exportPath); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	destDB := filepath.Join(dir, "dest.db")
	st2, err := ImportDB(ctx, exportPath, destDB)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()

	jobs, err := st2.ListJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].Company != "Round" {
		t.Fatalf("unexpected jobs after import: %+v", jobs)
	}
	applied, err := st2.AppliedMigrations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) == 0 {
		t.Fatal("expected migrations preserved in export")
	}
}

func TestImportRejectsNonSQLite(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.db")
	if err := os.WriteFile(bad, []byte("not sqlite"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ImportDB(context.Background(), bad, filepath.Join(dir, "out.db"))
	if err == nil {
		t.Fatal("expected error for non-sqlite input")
	}
}
