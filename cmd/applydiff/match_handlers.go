package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/match"
)

func (s *server) handleMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/match/")
	jobID := strings.Trim(path, "/")
	if jobID == "" || strings.Contains(jobID, "/") {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.computeMatch(w, r, jobID)
	case http.MethodGet:
		s.getMatch(w, r, jobID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) computeMatch(w http.ResponseWriter, r *http.Request, jobID string) {
	j, err := s.store.GetJob(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doc, err := s.store.LoadMaster(r.Context())
	if err != nil {
		http.Error(w, "no master cv saved", http.StatusNotFound)
		return
	}
	res := match.Evaluate(match.Input{Job: *j, CV: doc})
	if err := s.store.SaveMatchResult(r.Context(), res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if j.Status == "discovered" {
		_ = s.store.UpdateJobStatus(r.Context(), jobID, "matched")
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *server) getMatch(w http.ResponseWriter, r *http.Request, jobID string) {
	res, err := s.store.LoadMatchResult(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "no match result", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *server) handleJobByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	rest := strings.Trim(path, "/")
	if rest == "" || strings.Contains(rest, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Status == "" {
		http.Error(w, "status required", http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateJobStatus(r.Context(), rest, body.Status); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": body.Status})
}
