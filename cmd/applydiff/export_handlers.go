package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pahlevikun/applydiff/internal/application"
	"github.com/pahlevikun/applydiff/internal/export"
)

func (s *server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/export/")
	jobID := strings.Trim(path, "/")
	if jobID == "" || strings.Contains(jobID, "/") {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	j, err := s.store.GetJob(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	branch, err := s.store.LoadBranch(r.Context(), jobID)
	if err != nil {
		http.Error(w, "branch cv: "+err.Error(), http.StatusBadRequest)
		return
	}
	facts, err := s.store.LoadFacts(r.Context())
	if err != nil {
		facts = s.factsFromFile()
	}
	outRoot := os.Getenv("EXPORT_ROOT")
	if outRoot == "" {
		outRoot = filepath.Join("exports")
	}
	res, err := export.Build(export.Options{
		OutRoot: outRoot,
		Job:     *j,
		Doc:     branch,
		Facts:   facts,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := s.store.CreateApplication(r.Context(), jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app, err := s.store.GetApplication(r.Context(), jobID); err == nil {
		if application.CanTransition(app.Status, application.StatusReady) {
			_, _ = s.store.TransitionApplication(r.Context(), jobID, application.StatusReady, "")
		}
	}
	if err := s.store.SetApplicationCVSnapshot(r.Context(), jobID, res.CVSnapshot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
