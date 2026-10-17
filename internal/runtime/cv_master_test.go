package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func TestSaveMasterCVWritesJSONAndTeX(t *testing.T) {
	dir := t.TempDir()
	masterJSON := filepath.Join(dir, "master_cv.json")
	masterTeX := filepath.Join(dir, "master_cv.tex")
	dbPath := filepath.Join(dir, "applydiff.db")
	cfgPath := filepath.Join(dir, "jobs", "config.yaml")
	t.Setenv("DATABASE_PATH", dbPath)
	t.Setenv("JOBS_CONFIG", cfgPath)
	t.Setenv("MASTER_CV_JSON", masterJSON)
	t.Setenv("MASTER_CV_TEX", masterTeX)

	ctx := context.Background()
	app, err := Open(ctx, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	doc, err := cv.LoadFile(filepath.Join("..", "..", "cv", "master_cv.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveMasterCV(ctx, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(masterJSON); err != nil {
		t.Fatalf("master json: %v", err)
	}
	if _, err := os.Stat(masterTeX); err != nil {
		t.Fatalf("master tex: %v", err)
	}
	load, err := app.LoadMasterCV(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if load.Source != MasterCVFromDB {
		t.Fatalf("source %q", load.Source)
	}
	if load.Document.Person != doc.Person {
		t.Fatalf("person mismatch")
	}
}
