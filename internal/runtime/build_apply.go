package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/pahlevikun/applydiff/internal/application"
	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/store"
)

// BuildPDF exports LaTeX/PDF for a job branch.
func (a *App) BuildPDF(ctx context.Context, key string, dryRun bool) (*export.Result, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	if err := RequireAllDecided(prop); err != nil {
		return nil, err
	}
	if dryRun {
		return &export.Result{Dir: "dry-run"}, nil
	}
	branch, err := a.Store.LoadBranch(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	facts := a.FactsOrFile(ctx)
	outRoot := os.Getenv("OUTPUT_ROOT")
	if outRoot == "" {
		outRoot = os.Getenv("EXPORT_ROOT")
	}
	if outRoot == "" {
		outRoot = "output"
	}
	res, err := export.Build(export.Options{
		OutRoot: outRoot,
		Job:     *j,
		Doc:     branch,
		Facts:   facts,
	})
	if err != nil {
		return nil, err
	}
	if !res.PDFCompiled && res.CompileNote != "" {
		return res, ErrLaTeX{Note: res.CompileNote}
	}
	if _, err := a.Store.CreateApplication(ctx, j.ID); err != nil {
		return nil, err
	}
	if app, err := a.Store.GetApplication(ctx, j.ID); err == nil {
		if application.CanTransition(app.Status, application.StatusReady) {
			_, _ = a.Store.TransitionApplication(ctx, j.ID, application.StatusReady, "")
		}
	}
	_ = a.Store.SetApplicationCVSnapshot(ctx, j.ID, res.CVSnapshot)
	return res, nil
}

// ApplyMode controls how submission is recorded.
type ApplyMode string

const (
	ApplyManual    ApplyMode = "manual"
	ApplyAssisted  ApplyMode = "assisted"
	ApplyAuto      ApplyMode = "auto"
)

// ApplyResult is returned by Apply.
type ApplyResult struct {
	JobID   string `json:"job_id"`
	URL     string `json:"url"`
	Mode    string `json:"mode"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Apply records submission intent. Auto-submit is never performed.
func (a *App) Apply(ctx context.Context, key string, mode ApplyMode, dryRun bool) (*ApplyResult, error) {
	if mode == "" {
		mode = ApplyManual
	}
	if mode == ApplyAuto {
		return nil, fmt.Errorf("auto-submit is disabled; use manual or assisted")
	}
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	if err := RequireAllDecided(prop); err != nil {
		return nil, err
	}
	res := &ApplyResult{
		JobID:  j.ID,
		URL:    j.URL,
		Mode:   string(mode),
		Status: application.StatusReady,
	}
	switch mode {
	case ApplyAssisted:
		res.Message = "Open the job URL and submit manually; PDF should be ready in output/"
	case ApplyManual:
		res.Message = "Marked submitted after you apply manually at the job URL"
	default:
		return nil, fmt.Errorf("unknown apply mode %q", mode)
	}
	if dryRun {
		return res, nil
	}
	if _, err := a.Store.CreateApplication(ctx, j.ID); err != nil {
		return nil, err
	}
	app, err := a.Store.GetApplication(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	if application.CanTransition(app.Status, application.StatusSubmitted) {
		app, err = a.Store.TransitionApplication(ctx, j.ID, application.StatusSubmitted, "")
		if err != nil {
			return nil, err
		}
	}
	res.Status = app.Status
	return res, nil
}

// ExportDB copies the sqlite file.
func (a *App) ExportDB(ctx context.Context, dest string, dryRun bool) error {
	if dryRun {
		return nil
	}
	return a.Store.ExportDB(ctx, dest)
}

// ImportDB replaces the database from a file.
func (a *App) ImportDB(ctx context.Context, src string, dryRun bool) error {
	if dryRun {
		return nil
	}
	st, err := store.ImportDB(ctx, src, a.DBPath)
	if err != nil {
		return err
	}
	st.Close()
	return nil
}
