package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/store"
)

func (s *server) handleDBExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmp, err := os.CreateTemp("", "applydiff-export-*.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := s.app.Store.ExportDB(r.Context(), tmpPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f, err := os.Open(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/x-sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="applydiff.db"`)
	_, _ = io.Copy(w, f)
}

func (s *server) handleDBImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.dbPath == "" {
		http.Error(w, "import requires a file-backed DATABASE_PATH", http.StatusBadRequest)
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.dbPath), "applydiff-import-*.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpPath := tmp.Name()
	if _, err := io.Copy(tmp, r.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.app.Store.Close(); err != nil {
		os.Remove(tmpPath)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	st, err := store.ImportDB(r.Context(), tmpPath, s.dbPath)
	os.Remove(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.app.Store = st
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"imported"}`))
}
