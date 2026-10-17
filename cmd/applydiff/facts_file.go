package main

import (
	"context"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func (s *server) factsFromFile() *cv.Facts {
	f := s.app.FactsOrFile(context.Background())
	if f != nil {
		return f
	}
	return &cv.Facts{Facts: map[string]string{}}
}
