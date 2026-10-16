package main

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		view := s.cfg.BuildView(r.Context(), func(ctx context.Context, key string) (string, error) {
			return s.store.GetSetting(ctx, key)
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(view)
	case http.MethodPut:
		s.putConfigSources(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) putConfigSources(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SourceID string `json:"source_id"`
		Enabled  bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.SourceID == "" {
		http.Error(w, "source_id required", http.StatusBadRequest)
		return
	}
	view := s.cfg.BuildView(r.Context(), func(ctx context.Context, key string) (string, error) {
		return s.store.GetSetting(ctx, key)
	})
	overrides := map[string]bool{}
	for _, src := range view.Sources {
		overrides[src.ID] = src.Enabled
	}
	overrides[body.SourceID] = body.Enabled
	raw, err := json.Marshal(overrides)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.store.SetSetting(r.Context(), "source_enabled", string(raw)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if body.SourceID == "greenhouse" {
		s.cfg.Sources.Greenhouse.Enabled = body.Enabled
	}
	view = s.cfg.BuildView(r.Context(), func(ctx context.Context, key string) (string, error) {
		return s.store.GetSetting(ctx, key)
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(view)
}
