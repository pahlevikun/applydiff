package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/match"
)

func (s *server) handleMatch(w http.ResponseWriter, r *http.Request) {
	jobID := jobIDFromAPIPath("/api/match/", r.URL.Path)
	if jobID == "" {
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
	j, err := s.app.Store.GetJob(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doc, err := s.app.MasterCV(r.Context())
	if err != nil {
		http.Error(w, "no master cv saved", http.StatusNotFound)
		return
	}
	res := match.Evaluate(match.Input{Job: *j, CV: doc})
	if err := s.app.Store.SaveMatchResult(r.Context(), res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if j.Status == "discovered" {
		_ = s.app.Store.UpdateJobStatus(r.Context(), jobID, "matched")
	}
	_ = s.app.RecordMatchPipeline(r.Context(), jobID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *server) getMatch(w http.ResponseWriter, r *http.Request, jobID string) {
	res, err := s.app.Store.LoadMatchResult(r.Context(), jobID)
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
	rest := jobIDFromAPIPath("/api/jobs/", r.URL.Path)
	if rest == "" {
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
	appStatus := body.Status
	if appStatus == "discarded" {
		appStatus = "closed"
	}
	if err := s.app.SetJobDecision(r.Context(), rest, appStatus); err != nil {
		if strings.Contains(err.Error(), "not found") || err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "illegal transition") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": appStatus, "job_status": body.Status})
}
