package runtime

import (
	"context"

	"github.com/pahlevikun/applydiff/internal/jobconfig"
)

// SourcesConfigView builds the sources settings payload for web and TUI.
func (a *App) SourcesConfigView(ctx context.Context) jobconfig.ConfigView {
	if a == nil || a.Config == nil {
		return jobconfig.ConfigView{}
	}
	last, _ := a.Store.LoadSourceScanResults(ctx)
	view := a.Config.BuildView(ctx, func(ctx context.Context, key string) (string, error) {
		return a.Store.GetSetting(ctx, key)
	}, last)
	view.ConfigPath = a.CfgPath
	return view
}
