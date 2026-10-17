package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/filter"
)

type filtersPayload struct {
	Spec              filter.Spec `json:"spec"`
	ActivePresetID    string      `json:"active_preset_id"`
	ActivePresetName  string      `json:"active_preset_name,omitempty"`
}

func (s *server) handleFilters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getFilters(w, r)
	case http.MethodPut:
		s.putFilters(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) handleFilterPresets(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/filters/presets")
	path = strings.Trim(path, "/")
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			list, err := s.app.Store.ListFilterPresets(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, list)
		case http.MethodPost:
			var body struct {
				Name string       `json:"name"`
				Spec filter.Spec  `json:"spec"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if body.Name == "" {
				body.Name = "Preset"
			}
			if body.Spec.Arrangement == nil && body.Spec.Regions == nil {
				body.Spec = s.effectiveFilterSpec(r.Context())
			}
			p, err := s.app.Store.SaveFilterPreset(r.Context(), body.Name, body.Spec)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, p)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if strings.HasSuffix(path, "/apply") {
		id := strings.TrimSuffix(path, "/apply")
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := s.app.Store.SetActiveFilterPresetID(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]string{"active_preset_id": id})
		return
	}

	if r.Method == http.MethodDelete {
		if err := s.app.Store.DeleteFilterPreset(r.Context(), path); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.NotFound(w, r)
}

func (s *server) getFilters(w http.ResponseWriter, r *http.Request) {
	spec := s.effectiveFilterSpec(r.Context())
	activeID, _ := s.app.Store.GetActiveFilterPresetID(r.Context())
	payload := filtersPayload{Spec: spec, ActivePresetID: activeID}
	if activeID != "" {
		if p, err := s.app.Store.GetFilterPreset(r.Context(), activeID); err == nil {
			payload.ActivePresetName = p.Name
			payload.Spec = p.Spec
		}
	}
	writeJSON(w, payload)
}

func (s *server) putFilters(w http.ResponseWriter, r *http.Request) {
	var body filtersPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.app.Store.SaveFilterSpecOverride(r.Context(), body.Spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if body.ActivePresetID != "" {
		if err := s.app.Store.SetActiveFilterPresetID(r.Context(), body.ActivePresetID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	s.getFilters(w, r)
}

func (s *server) effectiveFilterSpec(ctx context.Context) filter.Spec {
	return s.app.EffectiveFilterSpec(ctx)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
