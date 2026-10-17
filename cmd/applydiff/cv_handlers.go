package main

import (
	"encoding/json"
	"net/http"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

type masterCVResponse struct {
	cv.Document
	Source     runtime.MasterCVSource `json:"_source,omitempty"`
	MasterJSON string                 `json:"_master_json,omitempty"`
	MasterTeX  string                 `json:"_master_tex,omitempty"`
}

func (s *server) handleCVMaster(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		load, err := s.app.LoadMasterCV(r.Context())
		if err != nil {
			http.Error(w, "no master cv saved", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(masterCVResponse{
			Document:   *load.Document,
			Source:     load.Source,
			MasterJSON: load.MasterJSON,
			MasterTeX:  load.MasterTeX,
		})
	case http.MethodPut:
		var doc cv.Document
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.app.SaveMasterCV(r.Context(), &doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":      "saved",
			"master_json": s.app.Paths.MasterJSON,
			"master_tex":  s.app.Paths.MasterTeX,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
