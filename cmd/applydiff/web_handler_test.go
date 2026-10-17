package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/webembed"
)

func TestEmbeddedUIHandler(t *testing.T) {
	sub, err := webembed.Sub()
	if err != nil {
		t.Skip("web assets not embedded")
	}
	_ = sub

	mux := http.NewServeMux()
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			return
		}
		serveEmbedded(w, r)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(strings.ToLower(body), "<html") {
		t.Fatalf("expected html body, got %q", truncateBody(body, 120))
	}
}

func truncateBody(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
