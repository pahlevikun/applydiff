package runtime_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestApplyConfigPatchQueries(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := &jobconfig.Config{Queries: []string{"a"}}
	if err := jobconfig.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     filepath.Join(dir, "test.db"),
		ConfigPath: cfgPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	next := []string{"golang", "remote"}
	if err := app.ApplyConfigPatch(ctx, runtime.ConfigPatch{Queries: &next}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := jobconfig.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Queries) != 2 || reloaded.Queries[0] != "golang" {
		t.Fatalf("queries %#v", reloaded.Queries)
	}
}

func TestApplyConfigPatchFilters(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := &jobconfig.Config{
		Filters: jobconfig.Filters{Spec: filter.Spec{RemoteOnly: true}},
	}
	if err := jobconfig.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	app, err := runtime.Open(ctx, runtime.Options{
		DBPath:     filepath.Join(dir, "test.db"),
		ConfigPath: cfgPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	spec := filter.Spec{
		RemoteOnly: false,
		Regions:    &filter.CategoryRule{Hard: true, Values: []string{"EU"}},
	}
	if err := app.ApplyConfigPatch(ctx, runtime.ConfigPatch{Filters: &spec}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := jobconfig.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Filters.Spec.Regions == nil || reloaded.Filters.Spec.Regions.Values[0] != "EU" {
		t.Fatalf("regions %#v", reloaded.Filters.Spec.Regions)
	}
}
