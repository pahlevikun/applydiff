package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pahlevikun/applydiff/internal/interview"
	"github.com/pahlevikun/applydiff/internal/runtime"
)

func (s *server) handleInterview(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/interview/")
	rest := strings.Trim(path, "/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(rest, "/")
	jobKey := parts[0]
	if jobKey == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mockType := r.URL.Query().Get("type")
		ws, err := s.app.InterviewWorkspace(r.Context(), jobKey, mockType)
		if err != nil {
			writeInterviewErr(w, err)
			return
		}
		writeJSON(w, ws)
		return
	}
	switch parts[1] {
	case "kit":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		kit, err := s.app.GenerateInterviewKit(r.Context(), jobKey)
		if err != nil {
			writeInterviewErr(w, err)
			return
		}
		writeJSON(w, kit)
	case "mock":
		if len(parts) < 3 || parts[2] != "save" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Type     string                    `json:"type"`
			Steps    []interview.MockStep      `json:"steps"`
			Scores   map[string]int            `json:"scores"`
			Comments map[string]string         `json:"comments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sess, err := s.app.SaveInterviewMock(r.Context(), jobKey, body.Type, body.Steps, body.Scores, body.Comments)
		if err != nil {
			writeInterviewErr(w, err)
			return
		}
		writeJSON(w, sess)
	case "stories":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var story interview.Story
		if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		saved, err := s.app.AddInterviewStory(r.Context(), jobKey, story)
		if err != nil {
			writeInterviewErr(w, err)
			return
		}
		writeJSON(w, saved)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) handleInterviewStory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/interview/stories/")
	id := strings.Trim(path, "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var story interview.Story
	if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	saved, err := s.app.UpdateInterviewStory(r.Context(), id, story)
	if err != nil {
		writeInterviewErr(w, err)
		return
	}
	writeJSON(w, saved)
}

func writeInterviewErr(w http.ResponseWriter, err error) {
	if _, ok := err.(runtime.ErrNoMatch); ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
