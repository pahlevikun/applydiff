package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/scan"
)

func runHTTPServer(port string, app *runtime.App) {
	if app == nil {
		log.Fatal("server: app not initialized")
	}
	srv := &server{app: app, dbPath: app.DBPath}

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
	mux.HandleFunc("/api/runs/", srv.handleRuns)
	mux.HandleFunc("/api/runs", srv.handleRuns)
	mux.HandleFunc("/api/pipeline", srv.handlePipeline)
	mux.HandleFunc("/api/jobs/", srv.handleJobByID)
	mux.HandleFunc("/api/interview/stories/", srv.handleInterviewStory)
	mux.HandleFunc("/api/interview/", srv.handleInterview)
	mountWebUI(mux)

	addr := ":" + port
	log.Printf("listening on %s (db=%s config=%s)", addr, app.DBPath, app.CfgPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	app    *runtime.App
	dbPath string
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
	jobs, err := s.app.ListJobs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, jobs)
}

func (s *server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	result, err := s.app.Runner.Run(ctx, scan.Options{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
