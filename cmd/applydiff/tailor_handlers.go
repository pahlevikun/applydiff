package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/tailor"
)

func (s *server) handleTailor(w http.ResponseWriter, r *http.Request) {
	jobID := jobIDFromAPIPath("/api/tailor/", r.URL.Path)
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		s.postTailor(w, r, jobID)
	case http.MethodGet:
		s.getTailor(w, r, jobID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) postTailor(w http.ResponseWriter, r *http.Request, jobID string) {
	if r.URL.Query().Get("async") == "1" {
		run, err := s.app.Runs.Start(r.Context(), "tailor", jobID, func(ctx context.Context) (any, error) {
			return s.app.Tailor(ctx, jobID, false)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(run)
		return
	}
	j, err := s.app.Store.GetJob(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	master, err := s.app.MasterCV(r.Context())
	if err != nil {
		http.Error(w, "no master cv saved", http.StatusNotFound)
		return
	}
	facts, err := s.app.Store.LoadFacts(r.Context())
	if err != nil {
		facts = s.factsFromFile()
	}
	eng := &tailor.Engine{Runner: s.app.LLM}
	hunks, err := eng.Generate(r.Context(), *j, master, facts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prop, err := s.app.Store.SaveProposal(r.Context(), jobID, hunks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = s.app.RecordTailorStarted(r.Context(), jobID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prop)
}

func (s *server) getTailor(w http.ResponseWriter, r *http.Request, jobID string) {
	prop, err := s.app.Store.LoadProposal(r.Context(), jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "no proposal", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prop)
}

func (s *server) handleHunkAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/hunks/")
	rest := strings.Trim(path, "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	hunkID, action := parts[0], parts[1]
	var status string
	switch action {
	case "accept":
		status = tailor.StatusAccepted
	case "reject":
		status = tailor.StatusRejected
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}
	h, err := s.app.Store.GetHunk(r.Context(), hunkID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "hunk not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.Status == tailor.StatusBlocked {
		http.Error(w, "hunk blocked by facts gate", http.StatusConflict)
		return
	}
	if err := s.app.Store.SetHunkStatus(r.Context(), hunkID, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jobID, err := s.app.Store.JobIDForHunk(r.Context(), hunkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prop, err := s.app.Store.LoadProposal(r.Context(), jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prop)
}
