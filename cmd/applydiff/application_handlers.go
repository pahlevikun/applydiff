package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func (s *server) handleApplicationsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items, err := s.app.Store.ListApplications(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (s *server) handleApplicationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/applications/")
	if !strings.HasSuffix(rest, "/status") {
		http.NotFound(w, r)
		return
	}
	jobID := strings.TrimSuffix(rest, "/status")
	jobID = strings.Trim(jobID, "/")
	if decoded, err := url.PathUnescape(jobID); err == nil {
		jobID = decoded
	}
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	var body struct {
		Status       string `json:"status"`
		FollowUpNote string `json:"follow_up_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := s.app.Store.GetJob(r.Context(), jobID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app, err := s.app.Store.TransitionApplication(r.Context(), jobID, body.Status, body.FollowUpNote)
	if err != nil {
		if strings.Contains(err.Error(), "illegal transition") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(app)
}
