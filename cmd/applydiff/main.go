package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/scan"
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
