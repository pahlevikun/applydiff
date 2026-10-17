package runtime

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/application"
)

// RecordMatchPipeline creates or advances the application row after a successful match score.
func (a *App) RecordMatchPipeline(ctx context.Context, jobID string) error {
	if a == nil || a.Store == nil || jobID == "" {
		return nil
	}
	app, err := a.Store.GetApplication(ctx, jobID)
	if err == sql.ErrNoRows {
		_, err = a.Store.TransitionApplication(ctx, jobID, application.StatusMatched, "")
		return err
	}
	if err != nil {
		return err
	}
	if app.Status == application.StatusDiscovered {
		_, err = a.Store.TransitionApplication(ctx, jobID, application.StatusMatched, "")
		return err
	}
	return nil
}

// RecordTailorStarted ensures an application row exists and marks tailoring in the pipeline.
func (a *App) RecordTailorStarted(ctx context.Context, jobID string) error {
	if a == nil || a.Store == nil || jobID == "" {
		return fmt.Errorf("app not ready")
	}
	_ = a.RecordMatchPipeline(ctx, jobID)
	_, err := a.Store.TransitionApplication(ctx, jobID, application.StatusTailoring, "")
	if err != nil {
		return err
	}
	return a.Store.UpdateJobStatus(ctx, jobID, "tailoring")
}

// SetJobDecision updates application pipeline status and mirrors it on jobs.status for the feed.
func (a *App) SetJobDecision(ctx context.Context, jobID, appStatus string) error {
	if a == nil || a.Store == nil {
		return fmt.Errorf("app not ready")
	}
	if jobID == "" || appStatus == "" {
		return fmt.Errorf("job id and status required")
	}
	_, err := a.Store.TransitionApplication(ctx, jobID, appStatus, "")
	if err != nil {
		return err
	}
	jobStatus := appStatus
	if appStatus == "closed" {
		jobStatus = "discarded"
	}
	return a.Store.UpdateJobStatus(ctx, jobID, jobStatus)
}
