package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/runs"
)

func (s *server) handleRuns(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/runs")
	path = strings.Trim(path, "/")
	if path == "" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		list, err := s.app.Store.ListRecentRuns(r.Context(), 40)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]runs.Run, 0, len(list))
		for _, row := range list {
			out = append(out, runs.Run{
				ID: row.ID, Kind: row.Kind, JobKey: row.JobKey, Status: row.Status,
				ResultJSON: row.ResultJSON, Error: row.Error, Created: row.Created, Updated: row.Updated,
			})
		}
		writeJSON(w, out)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	run, err := s.app.Runs.Get(r.Context(), path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(run)
}
