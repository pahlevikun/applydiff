package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/greenhouse"
	"github.com/pahlevikun/applydiff/internal/store"
)

type server struct {
	store  *store.Store
	cfg    *jobconfig.Config
	runner *scan.Runner
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "applydiff.db"
	}

	cfgPath := os.Getenv("JOBS_CONFIG")
	if cfgPath == "" {
		cfgPath = filepath.Join("jobs", "config.yaml")
	}

	cfg, err := jobconfig.Load(cfgPath)
	if err != nil {
		log.Fatalf("load jobs config: %v", err)
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	ctx := context.Background()
	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	registerSources(cfg)

	srv := &server{
		store: st,
		cfg:   cfg,
		runner: &scan.Runner{
			Store:  st,
			Config: cfg,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", srv.handleHealthz)
	mux.HandleFunc("/api/jobs", srv.handleJobs)
	mux.HandleFunc("/api/scan", srv.handleScan)

	addr := ":" + port
	log.Printf("listening on %s (db=%s config=%s)", addr, dbPath, cfgPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func registerSources(cfg *jobconfig.Config) {
	if cfg.Sources.Greenhouse.Enabled {
		source.Register(greenhouse.NewFromConfig(cfg.Sources.Greenhouse))
	}
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobs, err := s.store.ListJobs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jobs)
}

func (s *server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	result, err := s.runner.Run(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
