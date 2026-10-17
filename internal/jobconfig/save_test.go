package jobconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	orig := filepath.Join("..", "..", "jobs", "config.yaml")
	cfg, err := Load(orig)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Queries = append(cfg.Queries, "ui-edit-test")
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	again, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Queries) != len(cfg.Queries) {
		t.Fatalf("queries len %d vs %d", len(again.Queries), len(cfg.Queries))
	}
	if again.Queries[len(again.Queries)-1] != "ui-edit-test" {
		t.Fatalf("last query %q", again.Queries[len(again.Queries)-1])
	}
	_ = os.Remove(path)
}
