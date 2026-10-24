package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/runtime"
)

func (s *server) handleHunksList(w http.ResponseWriter, r *http.Request, jobKey string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sum, err := s.app.ListHunkReviews(r.Context(), jobKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "no proposal") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sum)
}

func (s *server) handleHunkDecisionPost(w http.ResponseWriter, r *http.Request, jobKey string, index int) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Decision string `json:"decision"`
		Text     string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sum, err := s.app.SetHunkDecision(r.Context(), jobKey, index, body.Decision, body.Text, false, runtime.ActorHuman)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "no proposal") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "out of range") || strings.Contains(err.Error(), "invalid decision") || strings.Contains(err.Error(), "skip requires") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.Contains(err.Error(), "blocked") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sum)
}

func (s *server) handleJobBuild(w http.ResponseWriter, r *http.Request, jobKey string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	res, err := s.app.BuildPDF(r.Context(), jobKey, false, runtime.ActorHuman)
	if err != nil {
		if strings.Contains(err.Error(), "still need approval") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}
