package source

import (
	"context"
	"time"
)

// RequestDelay is the pause between HTTP calls when a source issues multiple requests.
const RequestDelay = 250 * time.Millisecond

// PauseBetween waits politely between requests unless the context is cancelled.
func PauseBetween(ctx context.Context) {
	if RequestDelay <= 0 {
		return
	}
	t := time.NewTimer(RequestDelay)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
