package runtime_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestImportMasterCVExample(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     filepath.Join(dir, "test.db"),
		ConfigPath: filepath.Join("..", "..", "jobs", "config.yaml"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	master := filepath.Join("..", "..", "cv", "master_cv.json")
	if err := app.ImportMasterCV(ctx, master, false); err != nil {
		t.Fatal(err)
	}
	doc, err := app.Store.LoadMaster(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Person != "farhan-yuda-pahlevi" {
		t.Fatalf("unexpected person %q", doc.Person)
	}
	facts, err := app.Store.LoadFacts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if facts.Facts["exp.1.b4.abac"] != "6+ microservices ~40%" {
		t.Fatalf("missing demo facts: %#v", facts.Facts)
	}
}
