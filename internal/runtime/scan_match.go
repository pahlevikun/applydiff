package runtime

import (
	"context"
	"database/sql"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/match"
	"github.com/pahlevikun/applydiff/internal/scan"
)

// ScanOptions controls job discovery, risk filters, and presets.
type ScanOptions struct {
	SourceID string
	DryRun   bool
	Preset   string
	MinRisk  *int
	MaxRisk  *int
}

// Scan runs job sources.
func (a *App) Scan(ctx context.Context, opt ScanOptions) (scan.Result, error) {
	if a == nil || a.Runner == nil {
		return scan.Result{}, nil
	}
	if opt.DryRun {
		return a.Runner.Run(ctx, scan.Options{
			SourceID: opt.SourceID,
			DryRun:   true,
		})
	}
	a.Runner.MinRisk = opt.MinRisk
	a.Runner.MaxRisk = opt.MaxRisk
	a.Runner.Preset = opt.Preset
	return a.Runner.Run(ctx, scan.Options{
		SourceID: opt.SourceID,
		DryRun:   false,
	})
}

// ComputeMatch scores a job against the master CV.
func (a *App) ComputeMatch(ctx context.Context, key string, dryRun bool) (*match.Result, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	doc, err := a.Store.LoadMaster(ctx)
	if err != nil {
		return nil, err
	}
	res := match.Evaluate(match.Input{Job: *j, CV: doc})
	if res.Total < 8 {
		return &res, ErrNoMatch{Reason: "match score below threshold"}
	}
	if dryRun {
		return &res, nil
	}
	if err := a.Store.SaveMatchResult(ctx, res); err != nil {
		return nil, err
	}
	if j.Status == "discovered" {
		_ = a.Store.UpdateJobStatus(ctx, j.ID, "matched")
	}
	return &res, nil
}

// LoadMatch loads a stored match or computes one.
func (a *App) LoadMatch(ctx context.Context, key string) (*match.Result, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	res, err := a.Store.LoadMatchResult(ctx, j.ID)
	if err == sql.ErrNoRows {
		return a.ComputeMatch(ctx, key, false)
	}
	return res, err
}

// MasterCV loads the structured master document.
func (a *App) MasterCV(ctx context.Context) (*cv.Document, error) {
	return a.Store.LoadMaster(ctx)
}
