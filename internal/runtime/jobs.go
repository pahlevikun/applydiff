package runtime

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/match"
)

// ErrNoMatch is exit code 2.
type ErrNoMatch struct {
	Reason string
}

func (e ErrNoMatch) Error() string {
	if e.Reason == "" {
		return "no match"
	}
	return e.Reason
}

// ErrNeedsApproval is exit code 3.
type ErrNeedsApproval struct {
	Pending int
}

func (e ErrNeedsApproval) Error() string {
	return fmt.Sprintf("%d hunk(s) still need approval", e.Pending)
}

// ErrLLM is exit code 4.
type ErrLLM struct {
	Err error
}

func (e ErrLLM) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "llm error"
}

func (e ErrLLM) Unwrap() error { return e.Err }

// ErrLaTeX is exit code 5.
type ErrLaTeX struct {
	Note string
}

func (e ErrLaTeX) Error() string {
	if e.Note == "" {
		return "latex build failed"
	}
	return e.Note
}

// JobKey resolves a job id (exact match).
func (a *App) JobKey(ctx context.Context, key string) (*job.Job, error) {
	j, err := a.Store.GetJob(ctx, key)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found: %s", key)
	}
	return j, err
}

// EffectiveFilterSpec returns the active job filter spec.
func (a *App) EffectiveFilterSpec(ctx context.Context) filter.Spec {
	activeID, _ := a.Store.GetActiveFilterPresetID(ctx)
	if activeID != "" {
		if p, err := a.Store.GetFilterPreset(ctx, activeID); err == nil {
			return p.Spec
		}
	}
	if override, err := a.Store.LoadFilterSpecOverride(ctx); err == nil && override != nil {
		return *override
	}
	return a.Config.Filters.EffectiveSpec()
}

// ListJobs returns jobs with live filter evaluation.
func (a *App) ListJobs(ctx context.Context) ([]job.JobView, error) {
	jobs, err := a.Store.ListJobs(ctx)
	if err != nil {
		return nil, err
	}
	spec := a.EffectiveFilterSpec(ctx)
	out := make([]job.JobView, 0, len(jobs))
	for _, j := range jobs {
		peers := a.peerHashesExcluding(ctx, j.ID)
		a.assessAndPersistJobRisk(ctx, &j, peers)
		eval := filter.Evaluate(j, spec)
		if j.Status == "skipped" && j.FilterSkipReason != "" {
			eval.Keep = false
			if len(eval.Reasons) == 0 {
				eval.Reasons = []string{j.FilterSkipReason}
			}
		}
		var v job.JobView
		v.Job = j
		v.Filter.Keep = eval.Keep
		v.Filter.Score = eval.Score
		v.Filter.Reasons = eval.Reasons
		v.Priority = filter.Priority(j, eval.Score)
		out = append(out, v)
	}
	return out, nil
}

// ShowJob returns job detail plus match reasons when available.
type JobDetail struct {
	Job   job.Job       `json:"job"`
	Match *match.Result `json:"match,omitempty"`
	View  job.JobView   `json:"view"`
}

func (a *App) ShowJob(ctx context.Context, key string) (*JobDetail, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	views, err := a.ListJobs(ctx)
	if err != nil {
		return nil, err
	}
	var fv job.JobView
	for _, v := range views {
		if v.Job.ID == j.ID {
			fv = v
			break
		}
	}
	if fv.Job.ID == "" {
		fv.Job = *j
	}
	peers := a.peerHashesExcluding(ctx, j.ID)
	a.assessAndPersistJobRisk(ctx, j, peers)
	detail := &JobDetail{Job: *j, View: fv}
	res, err := a.Store.LoadMatchResult(ctx, j.ID)
	if err == nil {
		detail.Match = res
	}
	return detail, nil
}
