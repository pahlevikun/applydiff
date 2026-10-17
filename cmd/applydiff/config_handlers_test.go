package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/store"
)

func TestPutConfigQueriesOnly(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := jobconfig.Save(cfgPath, &jobconfig.Config{Queries: []string{"go"}}); err != nil {
		t.Fatal(err)
	}
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	cfg, err := jobconfig.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	srv := testServer(st, cfg, nil, nil)
	srv.app.CfgPath = cfgPath

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", srv.handleConfig)

	body := `{"queries":["golang","remote"]}`
	req := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put queries %d: %s", rec.Code, rec.Body.String())
	}
	reloaded, err := jobconfig.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Queries) != 2 || reloaded.Queries[0] != "golang" {
		t.Fatalf("queries %#v", reloaded.Queries)
	}
}
