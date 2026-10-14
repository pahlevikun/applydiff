package source

import (
	"context"
	"sync"

	"github.com/pahlevikun/applydiff/internal/job"
)

type Source interface {
	ID() string
	Fetch(ctx context.Context, queries []string) ([]job.Job, error)
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Source)
)

func Register(s Source) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[s.ID()] = s
}

func List() []Source {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Source, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	return out
}
