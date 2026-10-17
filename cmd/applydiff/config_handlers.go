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
	var body struct {
		SourceID *string       `json:"source_id"`
		Enabled  *bool         `json:"enabled"`
		Queries  *[]string     `json:"queries"`
		Filters  *filter.Spec  `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hasSource := body.SourceID != nil && *body.SourceID != ""
	hasQueries := body.Queries != nil
	hasFilters := body.Filters != nil
	if !hasSource && !hasQueries && !hasFilters {
		http.Error(w, "source_id, queries, or filters required", http.StatusBadRequest)
		return
	}
	patch := runtime.ConfigPatch{
		SourceID: body.SourceID,
		Enabled:  body.Enabled,
		Queries:  body.Queries,
		Filters:  body.Filters,
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
