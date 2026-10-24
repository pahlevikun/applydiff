package main

import (
	"net/http"
	"strconv"

	"github.com/pahlevikun/applydiff/internal/inbox"
)

func (s *server) handleInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	res, err := inbox.List(r.Context(), s.app.Store, inbox.ListOptions{
		State: r.URL.Query().Get("state"),
		Limit: limit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}
