package runtime

import (
	"context"

	"github.com/pahlevikun/applydiff/internal/application"
)

// ChannelStats returns per-channel application metrics from stored rows.
func (a *App) ChannelStats(ctx context.Context) ([]application.ChannelStats, error) {
	return a.Store.ChannelStats(ctx)
}
