package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/pahlevikun/applydiff/internal/webembed"
)

func mountWebUI(mux *http.ServeMux) {
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || strings.HasPrefix(r.URL.Path, "/api") {
			return
		}
		serveEmbedded(w, r)
	})
}

func serveEmbedded(w http.ResponseWriter, r *http.Request) {
	sub, err := webembed.Sub()
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(webembed.MissingMessage()))
		return
	}
	clean := path.Clean(r.URL.Path)
	if clean == "/" || clean == "." {
		clean = "/index.html"
	}
	name := strings.TrimPrefix(clean, "/")
	if name == "" {
		name = "index.html"
	}
	data, err := fs.ReadFile(sub, name)
	if err != nil {
		data, err = fs.ReadFile(sub, "index.html")
		if err != nil {
			data, err = fs.ReadFile(sub, "200.html")
		}
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(name, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(name, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(name, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
