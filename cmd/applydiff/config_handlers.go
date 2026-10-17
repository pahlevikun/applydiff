package main

import (
	"encoding/json"
	"net/http"

	"github.com/pahlevikun/applydiff/internal/features"
	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

type configAPIResponse struct {
	jobconfig.ConfigView
	Features features.View `json:"features"`
}

func encodeConfigView(w http.ResponseWriter, view jobconfig.ConfigView) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(configAPIResponse{
		ConfigView: view,
		Features:   features.APIView(),
	})
}

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		view := s.app.SourcesConfigView(r.Context())
		encodeConfigView(w, view)
	case http.MethodPut:
		s.putConfig(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) putConfig(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	patch := runtime.ConfigPatch{}
	if b, ok := raw["source_id"]; ok {
		var id string
		if err := json.Unmarshal(b, &id); err != nil {
			http.Error(w, "invalid source_id", http.StatusBadRequest)
			return
		}
		if id != "" {
			patch.SourceID = &id
		}
	}
	if b, ok := raw["enabled"]; ok && patch.SourceID != nil {
		var enabled bool
		if err := json.Unmarshal(b, &enabled); err != nil {
			http.Error(w, "invalid enabled", http.StatusBadRequest)
			return
		}
		patch.Enabled = &enabled
	}
	if b, ok := raw["queries"]; ok {
		var queries []string
		if err := json.Unmarshal(b, &queries); err != nil {
			http.Error(w, "queries must be a JSON array of strings", http.StatusBadRequest)
			return
		}
		patch.Queries = &queries
	}
	if b, ok := raw["filters"]; ok {
		var spec filter.Spec
		if err := json.Unmarshal(b, &spec); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		patch.Filters = &spec
	}
	hasSource := patch.SourceID != nil && *patch.SourceID != ""
	hasQueries := patch.Queries != nil
	hasFilters := patch.Filters != nil
	if !hasSource && !hasQueries && !hasFilters {
		http.Error(w, "send queries, filters, or source_id+enabled", http.StatusBadRequest)
		return
	}
	if hasSource && patch.Enabled == nil {
		http.Error(w, "enabled required when source_id is set", http.StatusBadRequest)
		return
	}
	if err := s.app.ApplyConfigPatch(r.Context(), patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	view := s.app.SourcesConfigView(r.Context())
	encodeConfigView(w, view)
}

func interviewPrepDisabled(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "interview prep disabled", http.StatusNotFound)
}
