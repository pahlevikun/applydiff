package main

import (
	"os"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func (s *server) factsFromFile() *cv.Facts {
	if s.facts != nil {
		return s.facts
	}
	path := os.Getenv("FACTS_PATH")
	if path == "" {
		path = "facts.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &cv.Facts{Facts: map[string]string{}}
	}
	f, err := cv.LoadFactsJSON(data)
	if err != nil {
		return &cv.Facts{Facts: map[string]string{}}
	}
	return f
}
