package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/greenhouse"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP API and embedded web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
	return cmd
}

func runServer() error {
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
		return err
	}

	st, err := store.New(dbPath)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := st.Migrate(ctx); err != nil {
		st.Close()
		return err
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
	mountWebUI(mux)

	addr := ":" + port
	log.Printf("listening on %s (db=%s config=%s)", addr, dbPath, cfgPath)
	return http.ListenAndServe(addr, mux)
}

func registerSources(cfg *jobconfig.Config) {
	if cfg.Sources.Greenhouse.Enabled {
		source.Register(greenhouse.NewFromConfig(cfg.Sources.Greenhouse))
	}
}
