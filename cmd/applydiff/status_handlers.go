package main

import (
	"net/http"
	"time"
)

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	st, err := s.app.Store.LoadSystemStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := map[string]string{}
	if st.LastScanAt != nil {
		payload["last_scan_at"] = st.LastScanAt.UTC().Format(time.RFC3339)
	}
	if st.DaemonSeenAt != nil {
		payload["daemon_seen_at"] = st.DaemonSeenAt.UTC().Format(time.RFC3339)
	}
	writeJSON(w, payload)
}
