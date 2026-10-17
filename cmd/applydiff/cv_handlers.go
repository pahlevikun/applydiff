package main

import (
	"encoding/json"
	"net/http"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func (s *server) handleCVMaster(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		doc, err := s.app.Store.LoadMaster(r.Context())
		if err != nil {
			http.Error(w, "no master cv saved", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	case http.MethodPut:
		var doc cv.Document
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.app.Store.SaveMaster(r.Context(), &doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
