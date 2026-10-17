package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/greenhouse"
	"github.com/pahlevikun/applydiff/internal/store"
)

type server struct {
	store  *store.Store
	dbPath string
	cfg    *jobconfig.Config
	runner *scan.Runner
	llm    llm.Runner
	facts  *cv.Facts
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "export-db":
			runExportDB(os.Args[2:])
			return
		case "import-db":
			runImportDB(os.Args[2:])
			return
		}
	}
	runServer()
}

func runServer() {
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

	llmRunner := llm.NewRunnerFromEnv()
	factsPath := os.Getenv("FACTS_PATH")
	if factsPath == "" {
		factsPath = "facts.json"
	}
	var facts *cv.Facts
	if data, err := os.ReadFile(factsPath); err == nil {
		facts, _ = cv.LoadFactsJSON(data)
	}

	srv := &server{
		store:  st,
		dbPath: dbPath,
		cfg:    cfg,
		runner: &scan.Runner{
			Store:  st,
			Config: cfg,
		},
		llm:   llmRunner,
		facts: facts,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", srv.handleHealthz)
	mux.HandleFunc("/api/jobs", srv.handleJobs)
	mux.HandleFunc("/api/filters", srv.handleFilters)
	mux.HandleFunc("/api/filters/presets", srv.handleFilterPresets)
	mux.HandleFunc("/api/filters/presets/", srv.handleFilterPresets)
	mux.HandleFunc("/api/scan", srv.handleScan)
	mux.HandleFunc("/api/cv/master", srv.handleCVMaster)
	mux.HandleFunc("/api/match/", srv.handleMatch)
	mux.HandleFunc("/api/tailor/", srv.handleTailor)
	mux.HandleFunc("/api/hunks/", srv.handleHunkAction)
	mux.HandleFunc("/api/applications", srv.handleApplicationsList)
	mux.HandleFunc("/api/applications/", srv.handleApplicationStatus)
	mux.HandleFunc("/api/export/", srv.handleExport)
	mux.HandleFunc("/api/db/export", srv.handleDBExport)
	mux.HandleFunc("/api/db/import", srv.handleDBImport)
	mux.HandleFunc("/api/config", srv.handleConfig)
	mux.HandleFunc("/api/jobs/", srv.handleJobByID)

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
	jobs, err := s.jobsWithFilters(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, jobs)
}

func (s *server) handleCVMaster(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		doc, err := s.store.LoadMaster(r.Context())
		if err != nil {
			http.Error(w, "no master cv saved", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	case http.MethodPut:
		var doc cv.Document
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.store.SaveMaster(r.Context(), &doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
